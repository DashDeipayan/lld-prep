package main

import "fmt"

type printGateway struct{}

func (printGateway) Charge(id string, amt int64) bool { fmt.Println("  charge", id, amt); return true }
func (printGateway) Refund(id string, amt int64)      { fmt.Println("  refund", id, amt) }

func main() {
	shop := NewOrderService(printGateway{})
	shop.AddProduct("KINDLE", 9999, 5)
	shop.AddProduct("CASE", 1499, 2)
	o1, _ := shop.PlaceOrder("asha", []OrderLine{{"KINDLE", 1}, {"CASE", 1}})
	total, _ := shop.Total(o1)
	fmt.Println(o1, "total", total, "stock KINDLE", shop.Stock("KINDLE"))
	shop.Pay(o1)
	if _, err := shop.PlaceOrder("ravi", []OrderLine{{"KINDLE", 1}, {"CASE", 2}}); err != nil {
		fmt.Println(err, "(KINDLE untouched:", shop.Stock("KINDLE"), ")")
	}
	shop.Cancel(o1)
	st, _ := shop.Status(o1)
	fmt.Println(o1, "status", st, "stock restored: CASE", shop.Stock("CASE"))
}
