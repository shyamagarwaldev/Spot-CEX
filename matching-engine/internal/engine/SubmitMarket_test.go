package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubmitMarket_BuyAgainstSingleAsk(t *testing.T) {
	// A market buy should consume the best available ask.
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 10, Ask, "seller", "A1", 1)

	fills, trades := ob.SubmitMarket(8, Bid, "buyer", "B1", 2)

	require.Len(t, fills, 2)
	require.Len(t, trades, 1)

	require.Equal(t, int64(8), fills[0].FilledQuantity)
	require.Equal(t, int64(0), fills[0].RemainingQuantity)
	require.Equal(t, Bid, fills[0].Side)

	require.Equal(t, "B1", trades[0].BuyOrderID)
	require.Equal(t, "A1", trades[0].SellOrderID)
	require.Equal(t, int64(100), trades[0].Price)
	require.Equal(t, int64(8), trades[0].Quantity)

	// Seller should have 2 units remaining.
	require.Equal(t, int64(2), ob.Orders["A1"].Quantity)
}

func TestSubmitMarket_BuyConsumesMultiplePriceLevels(t *testing.T) {
	// A market buy should walk through asks from the cheapest price upward.
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 5, Ask, "seller1", "A1", 1)
	ob.SubmitLimit(105, 5, Ask, "seller2", "A2", 2)
	ob.SubmitLimit(110, 5, Ask, "seller3", "A3", 3)

	fills, trades := ob.SubmitMarket(12, Bid, "buyer", "B1", 4)

	require.Len(t, trades, 3)

	require.Equal(t, int64(5), trades[0].Quantity)
	require.Equal(t, int64(100), trades[0].Price)
	require.Equal(t, "A1", trades[0].SellOrderID)

	require.Equal(t, int64(5), trades[1].Quantity)
	require.Equal(t, int64(105), trades[1].Price)
	require.Equal(t, "A2", trades[1].SellOrderID)

	require.Equal(t, int64(2), trades[2].Quantity)
	require.Equal(t, int64(110), trades[2].Price)
	require.Equal(t, "A3", trades[2].SellOrderID)

	require.Equal(t, int64(5*100+5*105+2*110), fills[0].TotalPrice)

	// A3 should have 3 remaining.
	require.Equal(t, int64(3), ob.Orders["A3"].Quantity)

	require.Equal(t, 0, ob.Bids.Size())
	require.Equal(t, 1, ob.Asks.Size())
}

func TestSubmitMarket_SellAgainstSingleBid(t *testing.T) {
	// A market sell should consume the best available bid.
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 10, Bid, "buyer", "B1", 1)

	fills, trades := ob.SubmitMarket(7, Ask, "seller", "S1", 2)

	require.Len(t, trades, 1)

	require.Equal(t, "B1", trades[0].BuyOrderID)
	require.Equal(t, "S1", trades[0].SellOrderID)
	require.Equal(t, int64(100), trades[0].Price)
	require.Equal(t, int64(7), trades[0].Quantity)

	require.Equal(t, int64(7), fills[0].FilledQuantity)
	require.Equal(t, int64(0), fills[0].RemainingQuantity)

	// Buyer should have 3 units remaining.
	require.Equal(t, int64(3), ob.Orders["B1"].Quantity)
}

func TestSubmitMarket_SellConsumesMultiplePriceLevels(t *testing.T) {
	// A market sell should walk through bids from the highest price downward.
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(110, 5, Bid, "buyer1", "B1", 1)
	ob.SubmitLimit(105, 5, Bid, "buyer2", "B2", 2)
	ob.SubmitLimit(100, 5, Bid, "buyer3", "B3", 3)

	fills, trades := ob.SubmitMarket(12, Ask, "seller", "S1", 4)

	require.Len(t, trades, 3)

	require.Equal(t, int64(110), trades[0].Price)
	require.Equal(t, int64(5), trades[0].Quantity)
	require.Equal(t, "B1", trades[0].BuyOrderID)

	require.Equal(t, int64(105), trades[1].Price)
	require.Equal(t, int64(5), trades[1].Quantity)
	require.Equal(t, "B2", trades[1].BuyOrderID)

	require.Equal(t, int64(100), trades[2].Price)
	require.Equal(t, int64(2), trades[2].Quantity)
	require.Equal(t, "B3", trades[2].BuyOrderID)

	require.Equal(t, int64(5*110+5*105+2*100), fills[0].TotalPrice)

	// Buyer should have 3 units remaining.
	require.Equal(t, int64(3), ob.Orders["B3"].Quantity)

	require.Equal(t, 1, ob.Bids.Size())
	require.Equal(t, 0, ob.Asks.Size())
}

func TestSubmitMarket_NoLiquidity(t *testing.T) {
	// A market order with no opposing liquidity should not create a book order.
	ob := NewOrderBook("BTC", "BITCOIN")

	fills, trades := ob.SubmitMarket(10, Bid, "buyer", "B1", 1)

	require.Len(t, trades, 0)
	require.Len(t, fills, 1)

	require.Equal(t, int64(0), fills[0].FilledQuantity)
	require.Equal(t, int64(10), fills[0].RemainingQuantity)

	require.Equal(t, 0, ob.Bids.Size())
	require.Equal(t, 0, ob.Asks.Size())
	require.NotContains(t, ob.Orders, "B1")
}

func TestSubmitMarket_PartiallyFilled(t *testing.T) {
	// If liquidity is insufficient, the market order should be partially filled and then expire.
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 5, Ask, "seller", "A1", 1)

	fills, trades := ob.SubmitMarket(10, Bid, "buyer", "B1", 2)

	require.Len(t, trades, 1)

	require.Equal(t, int64(5), trades[0].Quantity)
	require.Equal(t, int64(100), trades[0].Price)

	require.Equal(t, int64(5), fills[0].FilledQuantity)
	require.Equal(t, int64(5), fills[0].RemainingQuantity)

	// Market orders should not remain on the book.
	require.NotContains(t, ob.Orders, "B1")
	require.Equal(t, 0, ob.Bids.Size())
	require.Equal(t, 0, ob.Asks.Size())
}

func TestSubmitMarket_FIFOAtSamePrice(t *testing.T) {
	// A market order must respect FIFO among orders at the same price.
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 5, Ask, "seller1", "A1", 1)
	ob.SubmitLimit(100, 5, Ask, "seller2", "A2", 2)
	ob.SubmitLimit(100, 5, Ask, "seller3", "A3", 3)

	_, trades := ob.SubmitMarket(7, Bid, "buyer", "B1", 4)

	require.Len(t, trades, 2)

	require.Equal(t, "A1", trades[0].SellOrderID)
	require.Equal(t, int64(5), trades[0].Quantity)

	require.Equal(t, "A2", trades[1].SellOrderID)
	require.Equal(t, int64(2), trades[1].Quantity)

	// A2 should have 3 remaining.
	require.Equal(t, int64(3), ob.Orders["A2"].Quantity)

	// A3 should be untouched.
	require.Equal(t, int64(5), ob.Orders["A3"].Quantity)
}

func TestSubmitMarket_DoesNotCreateOrder(t *testing.T) {
	// Unlike a limit order, an unfilled market order must never enter the book.
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitMarket(10, Bid, "buyer", "M1", 1)

	require.Equal(t, 0, ob.Bids.Size())
	require.Equal(t, 0, ob.Asks.Size())
	require.NotContains(t, ob.Orders, "M1")
}

func TestSubmitMarket_QuantityConservation(t *testing.T) {
	// The total traded quantity must equal the quantity removed from the maker orders.
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 10, Ask, "seller1", "A1", 1)
	ob.SubmitLimit(105, 10, Ask, "seller2", "A2", 2)

	fills, trades := ob.SubmitMarket(15, Bid, "buyer", "B1", 3)

	var traded int64
	for _, trade := range trades {
		traded += trade.Quantity
	}

	require.Equal(t, fills[0].FilledQuantity, traded)
	require.Equal(t, int64(15), traded)

	require.Equal(t, int64(5), ob.Orders["A2"].Quantity)
}

func TestSubmitMarket_RequiredMarketFund_SingleAskLevel(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 10, Ask, "seller", "A1", 1)

	// Buying 7 BTC at the available ask price of 100
	required := ob.RequiredMarketFunds(7)

	require.Equal(t, int64(700), required)
}

func TestSubmitMarket_RequiredMarketFund_MultipleAskLevels(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 5, Ask, "seller1", "A1", 1)
	ob.SubmitLimit(105, 5, Ask, "seller2", "A2", 2)
	ob.SubmitLimit(110, 5, Ask, "seller3", "A3", 3)

	// Need 12 BTC:
	// 5 @ 100 = 500
	// 5 @ 105 = 525
	// 2 @ 110 = 220
	// Total = 1245 USDT
	required := ob.RequiredMarketFunds(12)

	require.Equal(t, int64(1245), required)
}

func TestSubmitMarket_RequiredMarketFund_InsufficientLiquidity(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 5, Ask, "seller", "A1", 1)

	// Only 5 BTC available, but requesting 10.
	required := ob.RequiredMarketFunds(10)

	// Function calculates the amount needed for the available
	// liquidity, not for the unavailable quantity.
	require.Equal(t, int64(500), required)
}
