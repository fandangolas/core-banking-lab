package events

import "bank-api/src/models"

// Broker manages client subscriptions and broadcasts transaction events.
type Broker struct {
	clients       map[chan models.TransactionEvent]bool
	newClients    chan chan models.TransactionEvent
	closedClients chan chan models.TransactionEvent
	events        chan models.TransactionEvent
}

// BrokerInstance is the global event broker.
var BrokerInstance = NewBroker()

// NewBroker creates and starts a new Broker.
func NewBroker() *Broker {
	b := &Broker{
		clients:       make(map[chan models.TransactionEvent]bool),
		newClients:    make(chan chan models.TransactionEvent),
		closedClients: make(chan chan models.TransactionEvent),
		events:        make(chan models.TransactionEvent),
	}

	go b.start()
	return b
}

func (b *Broker) start() {
	for {
		select {
		case client := <-b.newClients:
			b.clients[client] = true
		case client := <-b.closedClients:
			delete(b.clients, client)
			close(client)
		case event := <-b.events:
			for client := range b.clients {
				client <- event
			}
		}
	}
}

// Subscribe registers a new listener and returns its channel.
func (b *Broker) Subscribe() chan models.TransactionEvent {
	ch := make(chan models.TransactionEvent)
	b.newClients <- ch
	return ch
}

// Unsubscribe removes a listener.
func (b *Broker) Unsubscribe(ch chan models.TransactionEvent) {
	b.closedClients <- ch
}

// Publish sends the given event to all connected clients.
func (b *Broker) Publish(event models.TransactionEvent) {
	b.events <- event
}
