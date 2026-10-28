package main

import "fmt"

func main() {
	cart := NewCart()
	cart.Add("BOOK", 500, 3)
	cart.Add("PEN", 100, 1)
	fmt.Println("subtotal =", cart.Subtotal())

	promos := []DiscountStrategy{
		NewPercentageOff(10),
		NewFlatOff(300, 1500),
		NewBuyNGetOneFree("BOOK", 2),
	}
	for _, p := range promos {
		fmt.Printf("%T -> %d\n", p, p.Discount(cart))
	}
	fmt.Println("total =", NewCheckout(promos...).Total(cart))
}
