package main

import "fmt"

func main() {
	sw := NewSplitwise()
	users := []string{"alice", "bob", "carol", "dave"}
	for _, u := range users {
		sw.AddUser(u)
	}
	sw.AddExpense("alice", 120_00, users, EqualSplit{})
	sw.AddExpense("bob", 60_00, []string{"carol", "dave"}, NewExactSplit(map[string]int64{"carol": 40_00, "dave": 20_00}))
	sw.AddExpense("carol", 100_00, []string{"alice", "carol"}, NewPercentSplit(map[string]int{"alice": 70, "carol": 30}))
	for _, u := range users {
		fmt.Printf("%-6s balance %8.2f\n", u, float64(sw.Balance(u))/100)
	}
	fmt.Printf("simplified: %+v\n", sw.Simplify())
}
