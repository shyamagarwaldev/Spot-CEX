package ledger

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

type ILedgerRepository interface {
	Append(entry *LedgerEntry) error
	GetEntries(accountID string) ([]LedgerEntry, error)
	BulkAppend(entry ...*LedgerEntry) error
}

type LedgerRepository struct {
}

func NewLederRepository() *LedgerRepository

func (l *LedgerRepository) Append(entry *LedgerEntry) error
func (l *LedgerRepository) GetEntries(accountID string) ([]LedgerEntry, error)
func (l *LedgerRepository) BulkAppend(entry ...*LedgerEntry) error
