package main

import "fmt"

func main() {
	t := NewStockTicker()
	var alerts []string
	t.Subscribe(NewThresholdAlert("AAPL", 150, true, &alerts))
	stopLog := t.Subscribe(ListenerFunc(func(s string, p int64) { fmt.Println("tick", s, p) }))
	for _, p := range []int64{149, 151, 151, 155, 148, 150} {
		t.Update("AAPL", p)
	}
	stopLog()
	t.Update("AAPL", 140)
	fmt.Println("alerts =", alerts, "listeners =", t.ListenerCount())
}
