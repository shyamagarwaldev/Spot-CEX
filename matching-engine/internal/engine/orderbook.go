package engine

import (
	"github.com/emirpasic/gods/v2/maps/treemap"
)

type OrderType uint8

const (
	Market OrderType = iota
	Limit
)

type Side uint8

const (
	Bid Side = iota
	Ask
)

type Order struct {
	Sequence  uint64
	Quantity  int64
	Price     int64
	Side      Side
	OrderType OrderType
	ID        string
	UserID    string
	Prev      *Order
	Next      *Order
	level     *PriceLevel
}

type Fill struct {
	UserID            string
	OrderID           string
	Side              Side
	RemainingQuantity int64
	FilledQuantity    int64
	TotalPrice        int64
}

type Trade struct {
	ID          string
	BuyOrderID  string
	SellOrderID string

	BuyUserID  string
	SellUserID string

	Price     int64
	Quantity  int64
	TakerSide Side
}

type PriceLevel struct {
	Price int64

	Head *Order
	Tail *Order

	Quantity   int64
	OrderCount int
}

type OrderBook struct {
	Symbol string
	Name   string
	Bids   *treemap.Map[int64, *PriceLevel]
	Asks   *treemap.Map[int64, *PriceLevel]
	Orders map[string]*Order

	LastTradePrice    int64
	LastTradeQuantity int64
	LastOrderSequence uint64
}

func NewOrder(price, quantity int64, id, user_id string, side Side, order_type OrderType, sequence uint64) *Order {
	return &Order{
		Price:     price,
		Quantity:  quantity,
		ID:        id,
		UserID:    user_id,
		Side:      side,
		OrderType: order_type,
		Sequence:  sequence,
	}
}

func NewPriceLevel(price int64) *PriceLevel {
	return &PriceLevel{
		Price: price,
	}
}

func NewOrderBook(symbol, name string) *OrderBook {
	return &OrderBook{
		Name:   name,
		Symbol: symbol,
		Bids:   treemap.New[int64, *PriceLevel](),
		Asks:   treemap.New[int64, *PriceLevel](),
		Orders: make(map[string]*Order),
	}
}

func InsertOrder(mpp *treemap.Map[int64, *PriceLevel], incomingOrder *Order) {
	priceLevel, found := mpp.Get(incomingOrder.Price)
	if !found {
		priceLevel := NewPriceLevel(incomingOrder.Price)
		priceLevel.Head = incomingOrder
		priceLevel.Tail = incomingOrder
		incomingOrder.level = priceLevel
		priceLevel.Quantity += incomingOrder.Quantity
		priceLevel.OrderCount++
		mpp.Put(incomingOrder.Price, priceLevel)
		return
	}
	priceLevel.Tail.Next = incomingOrder
	incomingOrder.Prev = priceLevel.Tail
	incomingOrder.level = priceLevel
	priceLevel.Tail = incomingOrder
	priceLevel.Quantity += incomingOrder.Quantity
	priceLevel.OrderCount++
}

func (ob *OrderBook) insert(order *Order) {
	switch order.Side {
	case Bid:
		InsertOrder(ob.Bids, order)
	case Ask:
		InsertOrder(ob.Asks, order)
	default:
		return
	}
	ob.Orders[order.ID] = order
}

func (ob *OrderBook) Delete(id string) *Order {
	order, exist := ob.Orders[id]
	if !exist {
		return nil
	}
	priceLevel := order.level
	prev := order.Prev
	next := order.Next
	if next != nil {
		next.Prev = prev
	} else {
		priceLevel.Tail = prev
	}
	if prev != nil {
		prev.Next = next
	} else {
		priceLevel.Head = next
	}
	priceLevel.Quantity -= order.Quantity
	priceLevel.OrderCount--
	if priceLevel.OrderCount == 0 {
		if order.Side == Bid {
			ob.Bids.Remove(priceLevel.Price)
		} else {
			ob.Asks.Remove(priceLevel.Price)
		}
	}
	delete(ob.Orders, id)
	order.Next = nil
	order.Prev = nil
	order.level = nil
	return order
}

func (ob *OrderBook) partialDelete(id string, filledQuantity int64) *Order {
	order, exist := ob.Orders[id]
	if !exist {
		return nil
	}
	priceLevel := order.level
	priceLevel.Quantity -= filledQuantity
	order.Quantity -= filledQuantity
	return order
}

func (ob *OrderBook) BestBid() (int64, *PriceLevel, bool) {
	return ob.Bids.Max()
}

func (ob *OrderBook) BestAsk() (int64, *PriceLevel, bool) {
	return ob.Asks.Min()
}

func newFill(incomingOrder *Order) Fill {
	return Fill{
		OrderID:           incomingOrder.ID,
		UserID:            incomingOrder.UserID,
		Side:              incomingOrder.Side,
		RemainingQuantity: incomingOrder.Quantity,
		FilledQuantity:    0,
		TotalPrice:        0,
	}
}

func (ob *OrderBook) RequiredMarketFunds(quantity int64) int64 {
	var totalAmount int64 = 0

	it := ob.Asks.Iterator()

	for it.Next() && quantity > 0 {
		qty := min(quantity, it.Value().Quantity)
		totalAmount += it.Key() * qty
		quantity -= qty
	}
	return totalAmount
}

func (ob *OrderBook) matchBuy(incomingOrder *Order) ([]Fill, []Trade) {
	if _, ok := ob.Orders[incomingOrder.ID]; ok {
		return nil, nil
	}
	fills := []Fill{
		newFill(incomingOrder),
	}

	trades := []Trade{}

	for incomingOrder.Quantity > 0 {
		bestAskedPrice, level, exist := ob.BestAsk()
		if !exist || level.Head == nil || level.Tail == nil || level.OrderCount <= 0 {
			break
		}

		if incomingOrder.OrderType == Limit && bestAskedPrice > incomingOrder.Price {
			break
		}

		makerOrder := level.Head

		filled := min(makerOrder.Quantity, incomingOrder.Quantity)

		//Takers/buyer's fiils
		fills[0].FilledQuantity += filled
		fills[0].RemainingQuantity -= filled
		totalPrice := makerOrder.Price * filled
		fills[0].TotalPrice += totalPrice

		//Makers/seller's Fill
		currMaker := newFill(makerOrder)
		currMaker.FilledQuantity = filled
		currMaker.RemainingQuantity = makerOrder.Quantity - filled
		currMaker.TotalPrice = makerOrder.Price * filled

		trades = append(trades, Trade{
			BuyOrderID: incomingOrder.ID,
			BuyUserID:  incomingOrder.UserID,

			SellOrderID: makerOrder.ID,
			SellUserID:  makerOrder.UserID,

			Price:     makerOrder.Price,
			Quantity:  filled,
			TakerSide: incomingOrder.Side,
		})

		fills = append(fills, currMaker)

		incomingOrder.Quantity -= filled

		if makerOrder.Quantity == filled {
			ob.Delete(makerOrder.ID)
		} else {
			ob.partialDelete(makerOrder.ID, filled)
		}
	}

	if incomingOrder.OrderType == Limit && incomingOrder.Quantity > 0 {
		ob.insert(incomingOrder)
	}

	return fills, trades
}

func (ob *OrderBook) matchSell(incomingOrder *Order) ([]Fill, []Trade) {
	if _, ok := ob.Orders[incomingOrder.ID]; ok {
		return nil, nil
	}

	fills := []Fill{
		newFill(incomingOrder),
	}

	trades := []Trade{}

	for incomingOrder.Quantity > 0 {
		bestBidPrice, level, exist := ob.BestBid()
		if !exist || level.Head == nil || level.Tail == nil || level.OrderCount <= 0 {
			break
		}

		if incomingOrder.OrderType == Limit && bestBidPrice < incomingOrder.Price {
			break
		}

		makerOrder := level.Head

		filled := min(makerOrder.Quantity, incomingOrder.Quantity)

		//Takers/seller's fiils
		fills[0].FilledQuantity += filled
		fills[0].RemainingQuantity -= filled
		fills[0].TotalPrice += makerOrder.Price * filled

		//Makers/buyer's Fill
		currMaker := newFill(makerOrder)
		currMaker.FilledQuantity = filled
		currMaker.RemainingQuantity = makerOrder.Quantity - filled
		currMaker.TotalPrice = makerOrder.Price * filled

		trades = append(trades, Trade{
			SellOrderID: incomingOrder.ID,
			SellUserID:  incomingOrder.UserID,

			BuyOrderID: makerOrder.ID,
			BuyUserID:  makerOrder.UserID,

			Price:     makerOrder.Price,
			Quantity:  filled,
			TakerSide: incomingOrder.Side,
		})

		fills = append(fills, currMaker)

		incomingOrder.Quantity -= filled

		if makerOrder.Quantity == filled {
			ob.Delete(makerOrder.ID)
		} else {
			ob.partialDelete(makerOrder.ID, filled)
		}
	}

	if incomingOrder.OrderType == Limit && incomingOrder.Quantity > 0 {
		ob.insert(incomingOrder)
	}
	return fills, trades
}

func (ob *OrderBook) SubmitLimit(price, quantity int64, side Side, user_id, id string, sequence uint64) ([]Fill, []Trade) {
	order := NewOrder(price, quantity, id, user_id, side, Limit, sequence)
	switch order.Side {
	case Bid:
		return ob.matchBuy(order)
	case Ask:
		return ob.matchSell(order)
	default:
		return nil, nil
	}
}

func (ob *OrderBook) SubmitMarket(quantity int64, side Side, user_id, id string, sequence uint64) ([]Fill, []Trade) {
	order := NewOrder(0, quantity, id, user_id, side, Market, sequence)
	switch order.Side {
	case Bid:
		return ob.matchBuy(order)
	case Ask:
		return ob.matchSell(order)
	default:
		return nil, nil
	}
}
