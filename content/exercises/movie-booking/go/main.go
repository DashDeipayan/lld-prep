package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Unix(0, 0)
	gw := GatewayFunc(func(user string, amount int64) bool { return amount < 2000 })
	bms := NewBookingService(gw, func() time.Time { return now }, 5*time.Minute)
	bms.AddTheatre("pvr-koramangala", "Bengaluru")
	bms.AddScreen("pvr-koramangala", "audi-1", 2, 1, 4)
	bms.AddMovie("m1", "Interstellar", 169)
	show, _ := bms.AddShow("m1", "audi-1", now.Add(time.Hour), 250, 400)
	fmt.Println("shows in Bengaluru:", bms.FindShows("bengaluru", "m1"))
	h, _ := bms.Hold(show, "asha", []string{"C1", "C2", "A1"})
	seats, _ := bms.AvailableSeats(show)
	fmt.Println("held", h, "available now", len(seats))
	b, _ := bms.Confirm(h)
	fmt.Printf("%+v\n", *b)
	h2, _ := bms.Hold(show, "ravi", []string{"A2", "A3", "A4", "B1", "B2", "B3", "B4", "C3"})
	if _, err := bms.Confirm(h2); err != nil {
		seats, _ = bms.AvailableSeats(show)
		fmt.Println(err, "-> seats released, available", len(seats))
	}
	c, _ := bms.Cancel(b.ID)
	fmt.Printf("%+v\n", *c)
}
