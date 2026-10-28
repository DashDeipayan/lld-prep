package main

import "fmt"

func main() {
	g := NewGame(3)
	p := X
	for _, m := range [][2]int{{0, 0}, {1, 1}, {0, 1}, {2, 2}, {0, 2}} {
		s, err := g.Move(p, m[0], m[1])
		fmt.Printf("%v -> (%d,%d): status=%d err=%v\n", "XO"[p:p+1], m[0], m[1], s, err)
		p = 1 - p
	}
	for r := 0; r < 3; r++ {
		fmt.Println(string([]byte{g.Cell(r, 0), g.Cell(r, 1), g.Cell(r, 2)}))
	}
}
