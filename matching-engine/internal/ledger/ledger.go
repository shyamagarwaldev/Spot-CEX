package ledger

type LedgerEventType uint8

const (
	Deposit LedgerEventType = iota
	Withdrawal
	Reserve
	Release
	TradeDebit
	TradeCredit
)

type LedgerEvent struct {
	ID     string
	UserID string

	Asset  string
	Amount int64

	Type LedgerEventType

	ReferenceID string
}

type ILedgerRepository interface {
	Publish(*LedgerEvent) error
	BulkPublish(...*LedgerEvent) error
}
