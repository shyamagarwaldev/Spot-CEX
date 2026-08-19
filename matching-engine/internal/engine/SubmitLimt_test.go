package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// A limit order with no matching order should remain on the book.
func TestSubmitLimit_LimitOrder(t *testing.T) {
	orderBook := NewOrderBook("BTC", "BITCOIN")

	fills, trades := orderBook.SubmitLimit(100, 8, Bid, "1", "1", 1)

	require.Len(t, fills, 1)
	require.Len(t, trades, 0)

	takerFill := fills[0]

	require.Equal(t, int64(0), takerFill.FilledQuantity)
	require.Equal(t, int64(8), takerFill.RemainingQuantity)
	require.Equal(t, "1", takerFill.OrderID)
	require.Equal(t, Bid, takerFill.Side)
	require.Equal(t, int64(0), takerFill.TotalPrice)
	require.Equal(t, "1", takerFill.UserID)

	require.Equal(t, 1, orderBook.Bids.Size())
	require.Equal(t, 0, orderBook.Asks.Size())
}

// A marketable bid should match the best ask at the maker's price.
func TestSubmitLimit_SingleTrade_MarketableLimitOrder(t *testing.T) {
	orderBook := NewOrderBook("BTC", "BITCOIN")

	fills1, trades1 := orderBook.SubmitLimit(100, 8, Bid, "1", "1", 1)
	fills2, trades2 := orderBook.SubmitLimit(110, 8, Ask, "2", "2", 2)

	require.Len(t, fills1, 1)
	require.Len(t, fills2, 1)
	require.Len(t, trades1, 0)
	require.Len(t, trades2, 0)

	fills3, trades3 := orderBook.SubmitLimit(120, 8, Bid, "3", "3", 3)

	require.Len(t, fills3, 2)
	require.Len(t, trades3, 1)

	takerFill := fills3[0]
	makerFill := fills3[1]

	require.Equal(t, int64(8), takerFill.FilledQuantity)
	require.Equal(t, int64(0), takerFill.RemainingQuantity)
	require.Equal(t, Bid, takerFill.Side)
	require.Equal(t, int64(8*trades3[0].Price), takerFill.TotalPrice)

	require.Equal(t, int64(8), makerFill.FilledQuantity)
	require.Equal(t, int64(0), makerFill.RemainingQuantity)
	require.Equal(t, Ask, makerFill.Side)
	require.Equal(t, int64(8*trades3[0].Price), makerFill.TotalPrice)

	require.Equal(t, "3", trades3[0].BuyOrderID)
	require.Equal(t, "2", trades3[0].SellOrderID)
	require.Equal(t, int64(110), trades3[0].Price)
	require.Equal(t, int64(8), trades3[0].Quantity)
	require.Equal(t, Bid, trades3[0].TakerSide)

	require.Equal(t, 1, orderBook.Bids.Size())
	require.Equal(t, 0, orderBook.Asks.Size())
}

// A large marketable order should consume multiple price levels in order.
func TestSubmitLimit_MultiTrade_PartialLimitOrder(t *testing.T) {
	orderBook := NewOrderBook("BTC", "BITCOIN")

	fills1, trades1 := orderBook.SubmitLimit(100, 8, Ask, "1", "1", 1)
	fills2, trades2 := orderBook.SubmitLimit(100, 18, Ask, "2", "2", 2)
	fills3, trades3 := orderBook.SubmitLimit(110, 18, Ask, "3", "3", 3)

	require.Len(t, fills1, 1)
	require.Len(t, fills2, 1)
	require.Len(t, fills3, 1)
	require.Len(t, trades1, 0)
	require.Len(t, trades2, 0)
	require.Len(t, trades3, 0)

	fills4, trades4 := orderBook.SubmitLimit(110, 27, Bid, "4", "4", 4)

	// 100 * 26 + 110 * 1 = 2710
	require.Len(t, fills4, 4)
	require.Len(t, trades4, 3)

	takerFill := fills4[0]

	require.Equal(t, int64(27), takerFill.FilledQuantity)
	require.Equal(t, int64(0), takerFill.RemainingQuantity)
	require.Equal(t, Bid, takerFill.Side)
	require.Equal(t, int64(2710), takerFill.TotalPrice)

	require.Equal(t, "1", trades4[0].SellOrderID)
	require.Equal(t, int64(100), trades4[0].Price)
	require.Equal(t, int64(8), trades4[0].Quantity)

	require.Equal(t, "2", trades4[1].SellOrderID)
	require.Equal(t, int64(100), trades4[1].Price)
	require.Equal(t, int64(18), trades4[1].Quantity)

	require.Equal(t, "3", trades4[2].SellOrderID)
	require.Equal(t, int64(110), trades4[2].Price)
	require.Equal(t, int64(1), trades4[2].Quantity)

	require.Equal(t, int64(1), fills4[3].FilledQuantity)
	require.Equal(t, int64(17), fills4[3].RemainingQuantity)
	require.Equal(t, int64(110), fills4[3].TotalPrice)
	require.Equal(t, Ask, fills4[3].Side)

	require.Equal(t, 0, orderBook.Bids.Size())
	require.Equal(t, 1, orderBook.Asks.Size())
}

// Orders at the same price must be matched in FIFO order.
func TestSubmitLimit_FIFOAtSamePrice(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 10, Ask, "seller1", "A1", 1)
	ob.SubmitLimit(100, 10, Ask, "seller2", "A2", 2)

	_, trades := ob.SubmitLimit(100, 10, Bid, "buyer1", "B1", 3)

	require.Len(t, trades, 1)
	require.Equal(t, "A1", trades[0].SellOrderID)

	_, trades = ob.SubmitLimit(100, 10, Bid, "buyer2", "B2", 4)

	require.Len(t, trades, 1)
	require.Equal(t, "A2", trades[0].SellOrderID)
}

// An order should not match when its price does not cross the spread.
func TestSubmitLimit_NonMarketableOrder(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(110, 10, Ask, "seller", "A1", 1)

	fills, trades := ob.SubmitLimit(100, 10, Bid, "buyer", "B1", 2)

	require.Len(t, fills, 1)
	require.Len(t, trades, 0)

	require.Equal(t, int64(0), fills[0].FilledQuantity)
	require.Equal(t, int64(10), fills[0].RemainingQuantity)

	require.Equal(t, 1, ob.Bids.Size())
	require.Equal(t, 1, ob.Asks.Size())
}

// A partially filled maker should remain at the front of its price level.
func TestSubmitLimit_PartialMakerFill(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 20, Ask, "seller", "A1", 1)

	fills, trades := ob.SubmitLimit(100, 8, Bid, "buyer", "B1", 2)

	require.Len(t, fills, 2)
	require.Len(t, trades, 1)

	require.Equal(t, int64(8), fills[0].FilledQuantity)
	require.Equal(t, int64(0), fills[0].RemainingQuantity)

	require.Equal(t, int64(8), fills[1].FilledQuantity)
	require.Equal(t, int64(12), fills[1].RemainingQuantity)

	require.Equal(t, 1, ob.Asks.Size())
	require.Equal(t, 0, ob.Bids.Size())

	askLevel, exists := ob.Asks.Get(100)
	require.True(t, exists)
	require.Equal(t, int64(12), askLevel.Quantity)
	require.Equal(t, 1, askLevel.OrderCount)
	require.Equal(t, "A1", askLevel.Head.ID)
	require.Equal(t, "A1", askLevel.Tail.ID)
}

// A partially filled taker should be inserted back into the book.
func TestSubmitLimit_PartialTakerFill(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 5, Ask, "seller", "A1", 1)

	fills, trades := ob.SubmitLimit(110, 10, Bid, "buyer", "B1", 2)

	require.Len(t, fills, 2)
	require.Len(t, trades, 1)

	require.Equal(t, int64(5), fills[0].FilledQuantity)
	require.Equal(t, int64(5), fills[0].RemainingQuantity)

	require.Equal(t, int64(5), fills[1].FilledQuantity)
	require.Equal(t, int64(0), fills[1].RemainingQuantity)

	require.Equal(t, 1, ob.Bids.Size())
	require.Equal(t, 0, ob.Asks.Size())

	bidLevel, exists := ob.Bids.Get(110)
	require.True(t, exists)

	require.Equal(t, int64(5), bidLevel.Quantity)
	require.Equal(t, 1, bidLevel.OrderCount)
	require.Equal(t, "B1", bidLevel.Head.ID)
	require.Equal(t, "B1", bidLevel.Tail.ID)
}

// A Buyer should consume the best price before moving to a worse price.
func TestSubmitLimit_BuyBestPricePriority(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(105, 10, Ask, "seller1", "A1", 1)
	ob.SubmitLimit(100, 10, Ask, "seller2", "A2", 2)
	ob.SubmitLimit(110, 10, Ask, "seller3", "A3", 3)

	_, trades := ob.SubmitLimit(110, 15, Bid, "buyer", "B1", 4)

	require.Len(t, trades, 2)

	// The lowest ask must always be consumed first.
	require.Equal(t, "A2", trades[0].SellOrderID)
	require.Equal(t, int64(100), trades[0].Price)
	require.Equal(t, int64(10), trades[0].Quantity)

	require.Equal(t, "A1", trades[1].SellOrderID)
	require.Equal(t, int64(105), trades[1].Price)
	require.Equal(t, int64(5), trades[1].Quantity)

	askLevel, exists := ob.Asks.Get(105)
	require.True(t, exists)
	require.Equal(t, int64(5), askLevel.Quantity)

	_, exists = ob.Asks.Get(100)
	require.False(t, exists)
}

func TestSubmitLimit_DuplicateBuyOrder(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	fills, trades := ob.SubmitLimit(200, 15, Bid, "buyer", "B1", 1)

	require.Len(t, trades, 0)
	require.Len(t, fills, 1)

	fills, trades = ob.SubmitLimit(200, 15, Bid, "buyer", "B1", 2)

	require.Nil(t, fills)
	require.Nil(t, trades)

	bidLevel, exists := ob.Bids.Get(200)
	require.True(t, exists)
	require.Equal(t, int64(15), bidLevel.Quantity)
}

// A sell order should consume the highest bid before moving to a lower price.
func TestSubmitLimit_SellBestPricePriority(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(105, 10, Bid, "buyer1", "B1", 1)
	ob.SubmitLimit(110, 10, Bid, "buyer2", "B2", 2)
	ob.SubmitLimit(100, 10, Bid, "buyer3", "B3", 3)

	_, trades := ob.SubmitLimit(100, 15, Ask, "seller", "A1", 4)

	require.Len(t, trades, 2)

	require.Equal(t, "B2", trades[0].BuyOrderID)
	require.Equal(t, int64(110), trades[0].Price)
	require.Equal(t, int64(10), trades[0].Quantity)

	require.Equal(t, "B1", trades[1].BuyOrderID)
	require.Equal(t, int64(105), trades[1].Price)
	require.Equal(t, int64(5), trades[1].Quantity)

	bidLevel, exists := ob.Bids.Get(105)
	require.True(t, exists)
	require.Equal(t, int64(5), bidLevel.Quantity)

	_, exists = ob.Bids.Get(110)
	require.False(t, exists)
}

func TestSubmitLimit_DuplicateSellOrder(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	fills, trades := ob.SubmitLimit(200, 15, Ask, "seller", "A1", 1)

	require.Len(t, trades, 0)
	require.Len(t, fills, 1)

	fills, trades = ob.SubmitLimit(200, 15, Ask, "seller", "A1", 2)

	require.Nil(t, fills)
	require.Nil(t, trades)

	askLevel, exists := ob.Asks.Get(200)
	require.True(t, exists)
	require.Equal(t, int64(15), askLevel.Quantity)
}

// A fully consumed price level should be removed from the order book.
func TestSubmitLimit_RemoveEmptyPriceLevel(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 10, Ask, "seller", "A1", 1)

	_, trades := ob.SubmitLimit(100, 10, Bid, "buyer", "B1", 2)

	require.Len(t, trades, 1)

	require.Equal(t, 0, ob.Asks.Size())

	_, exists := ob.Asks.Get(100)
	require.False(t, exists)
}

// Multiple orders at different prices should preserve FIFO within each price level.
func TestSubmitLimit_FIFOIndependentPerPriceLevel(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 10, Ask, "seller1", "A1", 1)
	ob.SubmitLimit(100, 10, Ask, "seller2", "A2", 2)
	ob.SubmitLimit(105, 10, Ask, "seller3", "A3", 3)
	ob.SubmitLimit(105, 10, Ask, "seller4", "A4", 4)

	_, trades := ob.SubmitLimit(110, 25, Bid, "buyer", "B1", 5)

	require.Len(t, trades, 3)

	// Price priority first, FIFO second.
	require.Equal(t, "A1", trades[0].SellOrderID)
	require.Equal(t, "A2", trades[1].SellOrderID)
	require.Equal(t, "A3", trades[2].SellOrderID)

	require.Equal(t, int64(100), trades[0].Price)
	require.Equal(t, int64(100), trades[1].Price)
	require.Equal(t, int64(105), trades[2].Price)

	// A4 should remain untouched.
	level, exists := ob.Asks.Get(105)
	require.True(t, exists)
	require.Equal(t, int64(15), level.Quantity)
	require.Equal(t, 2, level.OrderCount)
	require.Equal(t, "A3", level.Head.ID)
	require.Equal(t, "A4", level.Tail.ID)
}

// A zero quantity order should not create a resting order.
func TestSubmitLimit_ZeroQuantity(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	fills, trades := ob.SubmitLimit(100, 0, Bid, "buyer", "B1", 1)

	require.Len(t, fills, 1)
	require.Len(t, trades, 0)

	require.Equal(t, int64(0), fills[0].FilledQuantity)
	require.Equal(t, int64(0), fills[0].RemainingQuantity)

	require.Equal(t, 0, ob.Bids.Size())
	require.Equal(t, 0, ob.Asks.Size())
}

// Trade information should correctly identify the buyer, seller, price and taker.
func TestSubmitLimit_TradeMetadata(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 10, Ask, "seller", "S1", 2)

	_, trades := ob.SubmitLimit(105, 4, Bid, "buyer", "B1", 3)

	require.Len(t, trades, 1)

	trade := trades[0]

	require.Equal(t, "B1", trade.BuyOrderID)
	require.Equal(t, "buyer", trade.BuyUserID)

	require.Equal(t, "S1", trade.SellOrderID)
	require.Equal(t, "seller", trade.SellUserID)

	require.Equal(t, int64(100), trade.Price)
	require.Equal(t, int64(4), trade.Quantity)
	require.Equal(t, Bid, trade.TakerSide)
}
