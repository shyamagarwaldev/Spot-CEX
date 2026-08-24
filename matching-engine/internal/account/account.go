package account

import (
	"fmt"

	"github.com/shyamagarwaldev/Spot-CEX/matching-engine/internal/ledger"
)

type AccountService struct {
	ledger   ledger.ILedgerRepository
	balances IBalanceStore
}

func NewAccountService(
	ledger ledger.ILedgerRepository,
	balances IBalanceStore,
) *AccountService {
	return &AccountService{
		ledger:   ledger,
		balances: balances,
	}
}

func (acc *AccountService) GetBalance(userID, asset string) (*Balance, error) {
	balance, ok := acc.balances.Get(userID, asset)
	if !ok {
		return nil, fmt.Errorf("balance not found for userID %v and asset %v", userID, asset)
	}
	return &balance, nil
}

func (acc *AccountService) Deposit(userID, asset string, amount int64) error {
	err := acc.balances.Update(userID, asset, func(b *Balance) error {
		b.Available += amount
		return nil
	})

	if err != nil {
		return fmt.Errorf("unable to deposit for userID: %v and error: %w", userID, err)
	}

	ledgerEvent := &ledger.LedgerEvent{
		UserID: userID,
		Asset:  asset,
		Amount: amount,
		Type:   ledger.Deposit,
	}

	err = acc.ledger.Publish(ledgerEvent)

	if err != nil {
		return fmt.Errorf("unbale to publish ledger event for deposit for userID; %v and error: %w", userID, err)
	}
	return nil
}

func (acc *AccountService) Withdraw(userID, asset string, amount int64) error {
	err := acc.balances.Update(userID, asset, func(b *Balance) error {
		if b.Available < amount {
			return fmt.Errorf("insufficient available asset: %v", b.Asset)
		}

		b.Available -= amount
		return nil
	})

	if err != nil {
		return fmt.Errorf("unable to withdraw for userID: %v and error: %w", userID, err)
	}

	ledgerEvent := &ledger.LedgerEvent{
		UserID: userID,
		Asset:  asset,
		Amount: amount,
		Type:   ledger.Withdrawal,
	}

	err = acc.ledger.Publish(ledgerEvent)

	if err != nil {
		return fmt.Errorf("unbale to publish ledger event for withdraw for userID; %v and error: %w", userID, err)
	}
	return nil
}

func (acc *AccountService) Reserve(
	userID string,
	asset string,
	amount int64,
	refrenceID string,
) error {
	err := acc.balances.Update(userID, asset, func(b *Balance) error {
		if b.Available < amount {
			return fmt.Errorf("insufficient available asset: %v", b.Asset)
		}

		b.Available -= amount
		b.Reserved += amount
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to reserve for userID: %v and error: %w", userID, err)
	}

	ledgerEvent := &ledger.LedgerEvent{
		UserID:      userID,
		Asset:       asset,
		Amount:      amount,
		Type:        ledger.Reserve,
		ReferenceID: refrenceID,
	}

	err = acc.ledger.Publish(ledgerEvent)

	if err != nil {
		return fmt.Errorf("unbale to publish ledger event for reserve for userID; %v and error: %w", userID, err)
	}
	return nil
}

func (acc *AccountService) Release(
	userID string,
	asset string,
	amount int64,
	refrenceID string,
) error {
	err := acc.balances.Update(userID, asset, func(b *Balance) error {
		if b.Reserved < amount {
			return fmt.Errorf("insufficient reserved asset: %v", b.Asset)
		}
		b.Reserved -= amount
		b.Available += amount
		return nil
	})

	if err != nil {
		return fmt.Errorf("unable to release for userID: %v and error: %w", userID, err)
	}

	ledgerEvent := &ledger.LedgerEvent{
		UserID:      userID,
		Asset:       asset,
		Amount:      amount,
		Type:        ledger.Release,
		ReferenceID: refrenceID,
	}

	err = acc.ledger.Publish(ledgerEvent)

	if err != nil {
		return fmt.Errorf("unbale to publish ledger event for release for userID; %v and error: %w", userID, err)
	}
	return nil
}

func (acc *AccountService) SettleTrade(settelment *Settlement) error {
	err := acc.balances.Transact(func(tx BalanceTransaction) error {
		err := tx.Update(settelment.SellerID, settelment.BaseAsset, func(b *Balance) error {

			if b.Reserved < settelment.Quantity {
				return fmt.Errorf("insufficient reserved asset: %v", b.Asset)
			}
			b.Reserved -= settelment.Quantity
			return nil
		})
		if err != nil {
			return fmt.Errorf("unable to settle trade for (userID: %v and asset: %v) with error: %w", settelment.SellerID, settelment.BaseAsset, err)
		}

		err = tx.Update(settelment.SellerID, settelment.QuoteAsset, func(b *Balance) error {
			amount := settelment.Quantity * settelment.Price
			b.Available += amount
			return nil
		})
		if err != nil {
			return fmt.Errorf("unable to settle trade for (userID: %v and asset: %v) with error: %w", settelment.SellerID, settelment.QuoteAsset, err)
		}

		err = tx.Update(settelment.BuyerID, settelment.BaseAsset, func(b *Balance) error {
			b.Available += settelment.Quantity
			return nil
		})
		if err != nil {
			return fmt.Errorf("unable to settle trade for (userID: %v and asset: %v) with error: %w", settelment.BuyerID, settelment.BaseAsset, err)
		}

		err = tx.Update(settelment.BuyerID, settelment.QuoteAsset, func(b *Balance) error {
			if b.Reserved < settelment.Quantity*settelment.Price {
				return fmt.Errorf("insufficient reserved asset: %v", b.Asset)
			}
			b.Reserved -= settelment.Quantity * settelment.Price
			return nil
		})
		if err != nil {
			return fmt.Errorf("unable to settle trade for (userID: %v and asset: %v) with error: %w", settelment.BuyerID, settelment.QuoteAsset, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("balances transaction failed error: %w", err)
	}
	ledgerEvent1 := &ledger.LedgerEvent{
		UserID:      settelment.SellerID,
		Asset:       settelment.BaseAsset,
		Amount:      settelment.Quantity,
		Type:        ledger.TradeDebit,
		ReferenceID: settelment.ReferenceID,
	}

	ledgerEvent2 := &ledger.LedgerEvent{
		UserID:      settelment.SellerID,
		Asset:       settelment.QuoteAsset,
		Amount:      settelment.Quantity * settelment.Price,
		Type:        ledger.TradeCredit,
		ReferenceID: settelment.ReferenceID,
	}

	ledgerEvent3 := &ledger.LedgerEvent{
		UserID:      settelment.BuyerID,
		Asset:       settelment.BaseAsset,
		Amount:      settelment.Quantity,
		Type:        ledger.TradeCredit,
		ReferenceID: settelment.ReferenceID,
	}

	ledgerEvent4 := &ledger.LedgerEvent{
		UserID:      settelment.BuyerID,
		Asset:       settelment.QuoteAsset,
		Amount:      settelment.Quantity * settelment.Price,
		Type:        ledger.TradeDebit,
		ReferenceID: settelment.ReferenceID,
	}
	err = acc.ledger.BulkPublish(ledgerEvent1, ledgerEvent2, ledgerEvent3, ledgerEvent4)

	if err != nil {
		return fmt.Errorf("unbale to bulk publish ledger events for trade settelment for (BuyerUserID: %v, SellerUserID: %v and asset: %v) with error: %w", settelment.BuyerID, settelment.SellerID, settelment.QuoteAsset, err)
	}
	return nil
}
