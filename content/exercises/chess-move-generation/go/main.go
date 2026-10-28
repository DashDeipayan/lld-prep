package main

import "fmt"

func main() {
	b := NewStandardBoard()
	m, _ := b.MovesFrom("b1")
	fmt.Println("b1 knight:", m)
	m, _ = b.MovesFrom("e2")
	fmt.Println("e2 pawn:  ", m)
	b.Move("e2", "e4")
	b.Move("d7", "d5")
	m, _ = b.MovesFrom("e4")
	fmt.Println("e4 pawn can capture:", m)
	b.Move("e4", "d5")
	fmt.Println("d5 now holds", b.PieceAt("d5"), "turn =", b.Turn())
	m, _ = b.MovesFrom("d1")
	fmt.Println("d1 queen: ", m)
}
