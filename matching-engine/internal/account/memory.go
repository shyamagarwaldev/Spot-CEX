package account

import (
	"fmt"
	"sync"
)

type MemoryBalanceStore struct {
	mu sync.RWMutex

	balances map[string]map[string]*Balance
}

func NewMemoryBalanceStore() *MemoryBalanceStore {
	return &MemoryBalanceStore{
		balances: make(map[string]map[string]*Balance),
	}
}

func (s *MemoryBalanceStore) Get(
	userID string,
	asset string,
) (Balance, bool) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	balance, ok := s.balances[userID][asset]
	if !ok {
		return Balance{}, false
	}

	return *balance, true
}

func (s *MemoryBalanceStore) Set(
	userID, asset string,
	balance Balance,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.balances[userID]; !ok {
		s.balances[userID] = make(map[string]*Balance)
	}
	s.balances[userID][asset] = &balance
}

func (s *MemoryBalanceStore) Update(
	userID string,
	asset string,
	fn func(*Balance) error,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	balance, ok := s.balances[userID][asset]
	if !ok {
		return fmt.Errorf("no balance found for userID: %v and asset: %v", userID, asset)
	}
	return fn(balance)
}

func cloneBalance(balances map[string]map[string]*Balance) map[string]map[string]*Balance {
	copy := make(map[string]map[string]*Balance)
	for userID, balance := range balances {
		copy[userID] = make(map[string]*Balance)
		for asset, b := range balance {
			copy[userID][asset] = &Balance{
				Asset:     b.Asset,
				Available: b.Available,
				Reserved:  b.Reserved,
			}
		}
	}
	return copy
}

func (s *MemoryBalanceStore) Transact(
	fn func(BalanceTransaction) error,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx := &memoryBalanceTransaction{
		balances: cloneBalance(s.balances),
	}
	if err := fn(tx); err != nil {
		return err
	}

	s.balances = tx.balances
	return nil
}

type memoryBalanceTransaction struct {
	balances map[string]map[string]*Balance
}

func (m *memoryBalanceTransaction) Update(
	userID string,
	asset string,
	fn func(*Balance) error,
) error {
	userBalances, ok := m.balances[userID]
	if !ok {
		return fmt.Errorf("user not found: %s", userID)
	}

	balance, ok := userBalances[asset]
	if !ok {
		return fmt.Errorf("asset not found: %s", asset)
	}

	return fn(balance)
}
