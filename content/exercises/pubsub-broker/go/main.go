package main

import "fmt"

func main() {
	b := NewBroker()
	b.CreateTopic("orders")
	b.Publish("orders", "o-1 created")
	b.Subscribe("orders", "billing", true)
	b.Subscribe("orders", "email", false)
	b.Publish("orders", "o-2 created")
	b.Publish("orders", "o-1 paid")
	m, _ := b.Poll("orders", "billing", 10)
	fmt.Println("billing:", m)
	m, _ = b.Poll("orders", "email", 10)
	fmt.Println("email:  ", m)
	lb, _ := b.Lag("orders", "billing")
	le, _ := b.Lag("orders", "email")
	fmt.Println("lags: billing =", lb, "email =", le)
	b.Seek("orders", "email", 0)
	m, _ = b.Poll("orders", "email", 2)
	fmt.Println("email after seek(0):", m)
}
