package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Unix(0, 0)
	lot := NewParkingLot(NewHourlyPricing(map[VehicleType]int64{Bike: 10, Car: 20, Truck: 50}),
		NearestSelector{}, func() time.Time { return now })
	lot.AddFloor(1, 1, 2, 1)
	lot.AddFloor(2, 2, 2, 1)
	board := NewDisplayBoard(lot)
	lot.AddListener(ListenerFunc(func(e Event) { fmt.Printf("  event: %+v\n", e) }))

	bike, _ := lot.Park(Vehicle{"KA-01-B", Bike})
	car, _ := lot.Park(Vehicle{"KA-02-C", Car})
	truck, _ := lot.Park(Vehicle{"KA-03-T", Truck})
	fmt.Printf("%+v\n%+v\n%+v\n", *bike, *car, *truck)
	fmt.Println("free medium (board) =", board.Free(Medium))

	now = now.Add(90 * time.Minute)
	r, _ := lot.Unpark(car.ID)
	fmt.Printf("%+v\n", *r)
	fmt.Println("free medium (board) =", board.Free(Medium))
}
