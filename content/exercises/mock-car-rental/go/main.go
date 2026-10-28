package main

import "fmt"

func main() {
	r := NewRentalService()
	r.AddBranch("blr-1", "Bengaluru")
	r.AddCar("blr-1", "KA01", Sedan, 2000)
	r.AddCar("blr-1", "KA02", Sedan, 1800)
	cars, _ := r.Search("bengaluru", Sedan, 10, 13)
	fmt.Println("search:", cars)
	res, _ := r.Reserve("asha", "KA02", 10, 13)
	fmt.Printf("%+v\n", res)
	cars, _ = r.Search("Bengaluru", Sedan, 12, 14)
	fmt.Println("search again:", cars)
	amt, _ := r.ReturnCar(res.ID, 14)
	fmt.Println("returned 1 day late, pays", amt)
}
