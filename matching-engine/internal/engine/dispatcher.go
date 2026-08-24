package engine

import (
	"fmt"
	"sync"
)

type IDispatcher interface {
	Dispatch(Command) error
}

type Dispatcher struct {
	mu      sync.RWMutex
	workers map[string]IOrderBookWorker
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		workers: make(map[string]IOrderBookWorker),
	}
}

func (d *Dispatcher) Register(symbol string, worker IOrderBookWorker) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.workers[symbol] = worker
}

func (d *Dispatcher) Dispatch(cmd Command) error {

	symbol, ok := GetSymbol(cmd)
	if !ok {
		return fmt.Errorf("command does not contain symbol")
	}

	d.mu.RLock()
	worker, ok := d.workers[symbol]
	d.mu.RUnlock()
	if !ok {
		return fmt.Errorf("no order book worker for symbol: %s", symbol)
	}

	worker.Submit(cmd)

	return nil
}

func GetSymbol(cmd Command) (string, bool) {
	switch p := cmd.Type; p {
	case CreateOrder:
		return cmd.Payload.(CreateOrderCommand).Symbol, true
	case CancelOrder:
		return cmd.Payload.(CancelOrderCommand).Symbol, true
	case GetDepth:
		return cmd.Payload.(GetDepthCommand).Symbol, true
	case GetTicker:
		return cmd.Payload.(GetTickerCommand).Symbol, true
	default:
		return "", false
	}
}
