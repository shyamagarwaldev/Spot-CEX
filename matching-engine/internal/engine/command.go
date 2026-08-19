package engine

type CommandType uint8

const (
	CreateOrder CommandType = iota
	CancelOrder
	Deposit
	Withdraw
	GetTicker
	GetDepth
)

type CommandResult struct {
	Response any
	Err      error
}

type Command struct {
	Sequence uint64
	Type     CommandType
	Payload  any
	Response chan CommandResult
}

type CreateOrderCommand struct {
	OrderID  string
	UserID   string
	Symbol   string
	Side     Side
	Type     OrderType
	Price    int64
	Quantity int64
	Asset    string
}

type CancelOrderCommand struct {
	OrderID string
	UserID  string
	Symbol  string
}

type GetTickerCommand struct {
	Symbol string
}

type GetDepthCommand struct {
	Symbol string
	Limit  int32
}
