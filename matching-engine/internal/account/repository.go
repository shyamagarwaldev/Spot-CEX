package account

type LedgerRepository interface {
	Append(entry *LedgerEntry) error
	GetEntries(accountID string) ([]LedgerEntry, error)
	BulkAppend(entry ...*LedgerEntry) error
}
