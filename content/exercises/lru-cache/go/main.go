package main

import "fmt"

func main() {
	c := NewLRU[string, int](2, func(k string, v int) { fmt.Printf("evicted %s=%d\n", k, v) })
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a")
	c.Put("c", 3)
	v, ok := c.Get("b")
	fmt.Println("keys", c.Keys(), "get b =", v, ok)
}
