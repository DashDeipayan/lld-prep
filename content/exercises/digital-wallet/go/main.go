package main

import "fmt"

func main() {
	w := NewWallet()
	w.OpenAccount("alice")
	w.OpenAccount("bob")
	w.Deposit("alice", 1000)
	id, _ := w.Transfer("req-1", "alice", "bob", 300)
	fmt.Println("transfer ->", id)
	id, _ = w.Transfer("req-1", "alice", "bob", 300)
	fmt.Println("retry    ->", id, "(same id, no double charge)")
	if _, err := w.Transfer("req-2", "bob", "alice", 5000); err != nil {
		fmt.Println("failed:", err)
	}
	a, _ := w.Balance("alice")
	b, _ := w.Balance("bob")
	fmt.Println("alice =", a, "bob =", b)
	h, _ := w.History("alice")
	fmt.Printf("alice history: %+v\n", h)
}
