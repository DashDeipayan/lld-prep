package main

import "fmt"

func main() {
	g := GetIDGenerator()
	fmt.Println(g.NextID())
	fmt.Println(g.NextID())
	fmt.Println(g.NextIDWithPrefix("ORD"))
	fmt.Println("same instance:", g == GetIDGenerator())
}
