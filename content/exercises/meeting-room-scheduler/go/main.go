package main

import "fmt"

func main() {
	s := NewScheduler()
	s.AddRoom("Huddle", 4)
	s.AddRoom("Board", 12)
	s.AddRoom("Focus", 2)
	b, _ := s.Book("asha", 600, 660, 3)
	fmt.Printf("%+v\n", b)
	b, _ = s.Book("ravi", 630, 690, 3)
	fmt.Printf("%+v\n", b)
	free, _ := s.FreeRooms(540, 600, 2)
	fmt.Println("free 9-10 for 2:", free)
	slots, _ := s.FreeSlots("Huddle", 540, 720)
	fmt.Println("Huddle free slots 9-12:", slots)
}
