package main

// ---- Provided: do not change Cart's API ----

type Line struct {
	SKU       string
	UnitPrice int64
	Qty       int
}

type Cart struct {
	lines map[string]*Line
	order []string
}

func NewCart() *Cart { return &Cart{lines: map[string]*Line{}} }

func (c *Cart) Add(sku string, unitPrice int64, qty int) {
	if l, ok := c.lines[sku]; ok {
		l.Qty += qty
		return
	}
	c.lines[sku] = &Line{SKU: sku, UnitPrice: unitPrice, Qty: qty}
	c.order = append(c.order, sku)
}

func (c *Cart) Line(sku string) (Line, bool) {
	l, ok := c.lines[sku]
	if !ok {
		return Line{}, false
	}
	return *l, true
}

func (c *Cart) Subtotal() int64 {
	var s int64
	for _, l := range c.lines {
		s += l.UnitPrice * int64(l.Qty)
	}
	return s
}

// ---- Your design ----

// DiscountStrategy returns the amount (cents) a promotion takes off the cart.
type DiscountStrategy interface {
	Discount(c *Cart) int64
}

func NewPercentageOff(percent int) DiscountStrategy {
	panic("TODO")
}

func NewFlatOff(amount, minSubtotal int64) DiscountStrategy {
	panic("TODO")
}

func NewBuyNGetOneFree(sku string, n int) DiscountStrategy {
	panic("TODO")
}

type Checkout struct {
	// TODO
}

func NewCheckout(strategies ...DiscountStrategy) *Checkout {
	panic("TODO")
}

func (co *Checkout) Total(c *Cart) int64 {
	panic("TODO")
}
