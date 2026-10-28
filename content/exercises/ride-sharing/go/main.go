package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Unix(0, 0)
	rates := map[VehicleClass]Rate{Mini: {50, 12, 2}, Sedan: {80, 15, 3}, SUV: {120, 20, 4}}
	uber := NewRideService(NewNearestDriver(10), NewStandardPricing(rates, 150), func() time.Time { return now })
	uber.AddDriver("d1", Mini, Location{0, 0}, 4.6)
	uber.AddDriver("d2", Mini, Location{2, 1}, 4.9)
	uber.SetOnline("d1", true)
	uber.SetOnline("d2", true)
	trip, _ := uber.RequestRide("asha", Location{3, 1}, Location{10, 5}, Mini)
	d, _ := uber.DriverOf(trip)
	fmt.Println(trip, "assigned to", d)
	uber.Arrived(trip)
	uber.Start(trip)
	now = now.Add(17*time.Minute + 5*time.Second)
	fare, _ := uber.End(trip)
	st, _ := uber.Status(trip)
	fmt.Println("fare (1.5x surge) =", fare, "status", st)
}
