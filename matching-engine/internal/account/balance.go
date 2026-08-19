package account

type Balance struct {
	Asset     string
	Available int64
	Reserved  int64
}

func (b *Balance) Total() int64 {
	return b.Available + b.Reserved
}

type BalanceTransaction interface {
	Update(
		userID, asset string,
		fn func(*Balance) error,
	) error
}

type IBalanceStore interface {
	Set(userID, asset string, balance Balance)
	Get(userID, asset string) (Balance, bool)
	Update(
		userID string,
		asset string,
		fn func(*Balance) error,
	) error

	Transact(
		fn func(BalanceTransaction) error,
	) error
}
