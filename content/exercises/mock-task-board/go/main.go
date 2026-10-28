package main

import "fmt"

func main() {
	b := NewTaskBoard()
	api, _ := b.Create("Design API", "pm", 4)
	ui, _ := b.Create("Build UI", "pm", 3)
	b.AddDependency(ui, api)
	b.Assign(api, "bob")
	b.Assign(ui, "carol")
	fmt.Println("blocked:", b.BlockedTasks())
	b.Move(api, InProgress, "bob")
	b.Move(api, InReview, "bob")
	b.Move(api, Done, "alice")
	b.Move(ui, InProgress, "carol")
	fmt.Println("blocked:", b.BlockedTasks())
	h, _ := b.History(api)
	fmt.Println("history", api+":", h)
}
