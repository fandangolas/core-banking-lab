package kafka

// Topic names for banking events
const (
	TopicAccountCreated        = "banking.accounts.created"
	TopicDepositRequests       = "banking.commands.deposit-requests"
	TopicTransactionDeposit    = "banking.transactions.deposit"
	TopicTransactionWithdrawal = "banking.transactions.withdrawal"
	TopicTransactionTransfer   = "banking.transactions.transfer"
	TopicTransactionFailed     = "banking.transactions.failed"
)

// GetAllTopics returns list of all topics
func GetAllTopics() []string {
	return []string{
		TopicAccountCreated,
		TopicDepositRequests,
		TopicTransactionDeposit,
		TopicTransactionWithdrawal,
		TopicTransactionTransfer,
		TopicTransactionFailed,
	}
}
