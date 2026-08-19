package engine

import (
	"fmt"

	"github.com/shyamagarwaldev/Spot-CEX/matching-engine/internal/account"
)

type IOrderBookWorker interface {
	Submit(Command)
}

type OrderBookWorker struct {
	book     *OrderBook
	commands chan Command
	account  account.IAccountService
}

func (w *OrderBookWorker) Run() {
	for cmd := range w.commands {
		w.handel(&cmd)
	}
}

func (w *OrderBookWorker) handel(cmd *Command) {
	var result CommandResult
	switch p := cmd.Type; p {
	case CreateOrder:
		result = w.handleCreateOrder(cmd)
	case CancelOrder:
	case GetTicker:
	case GetDepth:
	}

	cmd.Response <- result
}

func (w *OrderBookWorker) handleCreateOrder(cmd *Command) CommandResult {
	payload := cmd.Payload.(CreateOrderCommand)

	switch payload.Type {
	case Limit:
		return w.handleLimitOrder(cmd, payload)

	case Market:
		return w.handleMarketOrder(cmd, payload)

	default:
		return CommandResult{
			Err: fmt.Errorf("unsupported order type"),
		}
	}
}

func (w *OrderBookWorker) handleLimitOrder(
	cmd *Command,
	order CreateOrderCommand,
) CommandResult {

	if order.Side == Bid {
		amount := order.Price * order.Quantity

		// Reserve USDT
		if err := w.account.Reserve(
			order.UserID,
			"USDT",
			amount,
			order.OrderID,
		); err != nil {
			return CommandResult{Err: err}
		}

	} else {
		// Reserve base asset
		if err := w.account.Reserve(
			order.UserID,
			order.Asset, // depending on how you represent base asset
			order.Quantity,
			order.OrderID,
		); err != nil {
			return CommandResult{Err: err}
		}
	}

	fills, trades := w.book.SubmitLimit(order.Price, order.Quantity, order.Side, order.UserID, order.OrderID, cmd.Sequence)

	for _, trade := range trades {
		w.account.SettleTrade(&account.Settlement{
			TradeID:     trade.ID,
			BuyerID:     trade.BuyUserID,
			SellerID:    trade.SellUserID,
			BaseAsset:   order.Asset,
			QuoteAsset:  "USDT",
			Quantity:    trade.Quantity,
			Price:       trade.Price,
			ReferenceID: order.OrderID,
		})
	}

	if order.Side == Bid {
		takerFill := fills[0]
		amount := takerFill.FilledQuantity*order.Price - takerFill.TotalPrice
		if amount > 0 {
			if err := w.account.Release(
				order.UserID,
				"USDT",
				amount,
				order.OrderID,
			); err != nil {
				return CommandResult{Err: err}
			}
		}
	}
	return CommandResult{
		Response: fills[0],
	}
}

func (w *OrderBookWorker) handleMarketOrder(
	cmd *Command,
	order CreateOrderCommand,
) CommandResult {

	// 1. Reserve funds before touching the order book.
	var reserved int64
	var reserveAsset string

	if order.Side == Bid {
		// Market BUY:
		// We don't know the USDT requirement beforehand,
		// so calculate it from the current asks.
		reserved = w.book.RequiredMarketFunds(order.Quantity)
		reserveAsset = "USDT"

		if reserved == 0 {
			return CommandResult{
				Err: fmt.Errorf("insufficient market liquidity"),
			}
		}

	} else {
		// Market SELL:
		// We already know exactly how much base asset is required.
		reserved = order.Quantity
		reserveAsset = order.Asset
	}

	if err := w.account.Reserve(
		order.UserID,
		reserveAsset,
		reserved,
		order.OrderID,
	); err != nil {
		return CommandResult{Err: err}
	}

	// 2. Match the market order.
	fills, trades := w.book.SubmitMarket(
		order.Quantity,
		order.Side,
		order.UserID,
		order.OrderID,
		cmd.Sequence,
	)

	// 3. Settle every trade.
	for _, trade := range trades {
		if err := w.account.SettleTrade(&account.Settlement{
			TradeID:     trade.ID,
			BuyerID:     trade.BuyUserID,
			SellerID:    trade.SellUserID,
			BaseAsset:   order.Asset,
			QuoteAsset:  "USDT",
			Quantity:    trade.Quantity,
			Price:       trade.Price,
			ReferenceID: order.OrderID,
		}); err != nil {
			return CommandResult{Err: err}
		}
	}

	// 4. Release unused reservation.

	if order.Side == Ask {
		// For SELL, whatever wasn't filled is released.
		filledQuantity := fills[0].FilledQuantity

		unfilled := reserved - filledQuantity

		if unfilled > 0 {
			if err := w.account.Release(
				order.UserID,
				order.Asset,
				unfilled,
				order.OrderID,
			); err != nil {
				return CommandResult{Err: err}
			}
		}
	}

	return CommandResult{
		Response: fills[0],
	}
}
