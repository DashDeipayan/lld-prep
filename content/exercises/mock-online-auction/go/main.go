package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Unix(0, 0)
	h := NewAuctionHouse(func() time.Time { return now })
	a, _ := h.Create("seller", "Vintage Camera", 1000, 100, now.Add(10*time.Minute))
	h.Subscribe(a, ListenerFunc(func(id, e string) { fmt.Println(" ", id, e) }))
	h.PlaceBid(a, "asha", 1000)
	h.PlaceBid(a, "ravi", 1200)
	now = now.Add(9*time.Minute + 30*time.Second)
	h.PlaceBid(a, "asha", 1300)
	fmt.Println("anti-sniping extended end to", h.EndTime(a).Sub(time.Unix(0, 0)))
	now = h.EndTime(a)
	r, _ := h.Close(a)
	fmt.Printf("%+v\n", r)
}
