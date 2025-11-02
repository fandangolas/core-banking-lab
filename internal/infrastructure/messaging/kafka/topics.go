package kafka

// Topic names for banking events
const (
	TopicAccountCreated        = "banking.accounts.created"
	TopicTransactionDeposit    = "banking.transactions.deposit"
	TopicTransactionWithdrawal = "banking.transactions.withdrawal"
	TopicTransactionTransfer   = "banking.transactions.transfer"
	TopicTransactionFailed     = "banking.transactions.failed"
)

// GetAllTopics returns list of all topics
func GetAllTopics() []string {
	return []string{
		TopicAccountCreated,
		TopicTransactionDeposit,
		TopicTransactionWithdrawal,
		TopicTransactionTransfer,
		TopicTransactionFailed,
	}
}
