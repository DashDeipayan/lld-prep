package main

import "fmt"

func main() {
	h := NewHotel()
	h.AddRoom("101", Single, 3000)
	h.AddRoom("102", Double, 4500)
	h.AddRoom("201", Suite, 9000)
	av, _ := h.Availability(10, 12)
	fmt.Println("availability:", av)
	b, _ := h.Book("asha", Double, 10, 12)
	r, _ := h.RoomOf(b)
	fmt.Println(b, "-> room", r)
	h.CheckIn(b, 10)
	h.AddCharge(b, 750)
	bill, _ := h.CheckOut(b)
	st, _ := h.Status(b)
	fmt.Println("bill =", bill, "("+st+")")
}
