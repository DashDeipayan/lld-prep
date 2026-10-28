package main

type PriceListener interface {
	OnPrice(symbol string, price int64)
}

type StockTicker struct {
	// TODO
}

func NewStockTicker() *StockTicker { panic("TODO") }

// Subscribe returns a function that unsubscribes the listener.
func (t *StockTicker) Subscribe(l PriceListener) func() { panic("TODO") }

func (t *StockTicker) Update(symbol string, price int64) { panic("TODO") }

func (t *StockTicker) ListenerCount() int { panic("TODO") }

func NewThresholdAlert(symbol string, threshold int64, above bool, sink *[]string) PriceListener {
	panic("TODO")
}

// ListenerFunc lets a plain function be a listener (handy in tests and main).
type ListenerFunc func(symbol string, price int64)

func (f ListenerFunc) OnPrice(s string, p int64) { f(s, p) }
