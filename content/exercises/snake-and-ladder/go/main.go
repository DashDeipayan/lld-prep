package main

import (
	"fmt"
	"math/rand"
)

func main() {
	b, err := NewBoard(100, map[int]int{99: 7, 62: 19, 48: 26}, map[int]int{2: 38, 28: 84, 51: 67, 71: 91})
	if err != nil {
		fmt.Println(err)
		return
	}
	r := rand.New(rand.NewSource(42))
	g, _ := NewGame(b, []string{"alice", "bob"}, DiceFunc(func() int { return 1 + r.Intn(6) }))
	turns := 0
	for {
		if _, done := g.Winner(); done || turns >= 1000 {
			break
		}
		res, _ := g.PlayTurn()
		if turns < 8 {
			fmt.Printf("%+v\n", res)
		}
		turns++
	}
	w, _ := g.Winner()
	fmt.Println("... winner =", w, "after", turns, "rolls")
}
