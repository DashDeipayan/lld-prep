package main

import "fmt"

func main() {
	ed := NewTextEditor(100)
	ed.Insert(0, "hello")
	ed.Insert(5, " world")
	ed.Delete(0, 6)
	fmt.Printf("%q\n", ed.Text())
	ed.Undo()
	fmt.Printf("undo -> %q\n", ed.Text())
	ed.Undo()
	fmt.Printf("undo -> %q\n", ed.Text())
	ed.Redo()
	fmt.Printf("redo -> %q\n", ed.Text())
	ed.Insert(11, "!")
	fmt.Printf("edit -> %q, redo possible? %v\n", ed.Text(), ed.Redo())
}
