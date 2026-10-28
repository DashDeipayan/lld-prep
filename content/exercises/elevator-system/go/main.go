package main

import "fmt"

func main() {
	sys := NewElevatorSystem(10, 2, NearestCar{})
	id, _ := sys.RequestPickup(5, Up)
	fmt.Println("pickup 5 UP -> car", id)
	sys.Step()
	sys.Step()
	id, _ = sys.RequestPickup(4, Up)
	fmt.Println("pickup 4 UP -> car", id)
	id, _ = sys.RequestPickup(1, Up)
	fmt.Println("pickup 1 UP -> car", id)
	for t := 3; t <= 8; t++ {
		ev := sys.Step()
		fmt.Printf("tick %d: car0@%d car1@%d %v\n", t, sys.Car(0).Floor(), sys.Car(1).Floor(), ev)
	}
}
