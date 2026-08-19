package account

type Settlement struct {
	TradeID string

	BuyerID  string
	SellerID string

	BaseAsset  string
	QuoteAsset string

	Quantity int64
	Price    int64

	ReferenceID string
}

type IAccountService interface {
	GetBalance(userID, asset string) (*Balance, error)

	Deposit(
		userID string,
		asset string,
		amount int64,
	) error

	Withdraw(
		UserID string,
		assert string,
		amount int64,
	) error

	Reserve(
		userID string,
		asset string,
		amount int64,
		refrenceID string,
	) error

	Release(
		userID string,
		asset string,
		amount int64,
		refrenceID string,
	) error

	SettleTrade(settelment *Settlement) error
}
