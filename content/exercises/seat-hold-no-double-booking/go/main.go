package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Unix(0, 0)
	inv := NewSeatInventory([]string{"A1", "A2", "A3", "A4"}, 5*time.Minute, func() time.Time { return now })
	h1, _ := inv.Hold("alice", []string{"A1", "A2"})
	fmt.Println("alice holds", h1, "available", inv.Available())
	if _, err := inv.Hold("bob", []string{"A2", "A3"}); err != nil {
		fmt.Println("bob failed:", err, "(A3 untouched:", inv.Status("A3")+")")
	}
	now = now.Add(6 * time.Minute)
	fmt.Println("after expiry, available", inv.Available())
	h2, _ := inv.Hold("bob", []string{"A2", "A3"})
	inv.Confirm(h2)
	fmt.Println("bob booked: A2 =", inv.Status("A2"), "A3 =", inv.Status("A3"))
}
