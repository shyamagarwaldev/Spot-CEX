package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Cancelling a partially filled order should remove only its remaining quantity.
func TestDelete_PartiallyFilledOrder(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 20, Ask, "seller", "A1", 1)
	ob.SubmitLimit(100, 8, Bid, "buyer", "B1", 2)

	level, exists := ob.Asks.Get(100)
	require.True(t, exists)
	require.Equal(t, int64(12), level.Quantity)

	cancelled := ob.Delete("A1")

	require.NotNil(t, cancelled)
	require.Equal(t, int64(12), cancelled.Quantity)

	require.Equal(t, 0, ob.Asks.Size())
	require.Nil(t, cancelled.level)
	require.Nil(t, cancelled.Next)
	require.Nil(t, cancelled.Prev)
}

func TestDelete_RecentOrder(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	ob.SubmitLimit(100, 20, Ask, "seller1", "A1", 1)
	ob.SubmitLimit(100, 8, Ask, "seller2", "A2", 2)

	level, exists := ob.Asks.Get(100)
	require.True(t, exists)
	require.Equal(t, int64(28), level.Quantity)

	cancelled := ob.Delete("A2")

	require.NotNil(t, cancelled)
	require.Equal(t, int64(8), cancelled.Quantity)
	require.Equal(t, "A2", cancelled.ID)

	require.Equal(t, 1, ob.Asks.Size())
	require.Nil(t, cancelled.level)
	require.Nil(t, cancelled.Next)
	require.Nil(t, cancelled.Prev)
}

// Deleting an unknown order should be a no-op.
func TestDelete_UnknownOrder(t *testing.T) {
	ob := NewOrderBook("BTC", "BITCOIN")

	require.Nil(t, ob.Delete("does-not-exist"))
}
