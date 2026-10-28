package main

import "fmt"

func main() {
	f := NewDefaultFactory()
	fmt.Println("supported =", f.SupportedMethods())
	for _, m := range f.SupportedMethods() {
		p, _ := f.Create(m)
		r, err := p.Pay(10_000)
		fmt.Printf("%+v %v\n", r, err)
	}
	if _, err := f.Create("bitcoin"); err != nil {
		fmt.Println("error:", err)
	}
}
