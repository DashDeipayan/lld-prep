package main

import "fmt"

func main() {
	var console, alerts []string
	log := NewLogger("orders", Info).
		AddHandler(NewMemoryHandler(Debug, &console, SimpleFormatter{})).
		AddHandler(NewMemoryHandler(Error, &alerts, JSONFormatter{}))
	log.Info("created #1")
	log.Debug("cache miss")
	log.Error(`charge failed for "card"`)
	fmt.Println("console:", console)
	fmt.Println("alerts: ", alerts)
}
