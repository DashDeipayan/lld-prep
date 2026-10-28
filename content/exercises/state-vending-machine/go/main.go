package main

import "fmt"

func main() {
	vm := NewVendingMachine()
	fmt.Println("state:", vm.State())
	vm.Stock(Product{"A1", "Cola", 125}, 1)
	vm.Stock(Product{"B2", "Chips", 50}, 2)
	vm.InsertCoin(100)
	vm.InsertCoin(25)
	vm.InsertCoin(25)
	d, err := vm.Select("A1")
	fmt.Printf("%+v %v -> state %s\n", d, err, vm.State())
	if _, err := vm.Select("B2"); err != nil {
		fmt.Println("error:", err)
	}
	vm.InsertCoin(10)
	fmt.Println("refund", vm.Cancel(), "state", vm.State())
}
