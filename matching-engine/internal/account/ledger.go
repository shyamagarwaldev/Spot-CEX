package account

type LedgerEntryType uint8

const (
	Deposit LedgerEntryType = iota
	Withdrawal
	Reserve
	Release
	TradeDebit
	TradeCredit
)

type LedgerEntry struct {
	ID     string
	UserID string

	Asset  string
	Amount int64

	Type LedgerEntryType

	ReferenceID string
}
