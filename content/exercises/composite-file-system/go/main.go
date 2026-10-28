package main

import "fmt"

func main() {
	fs := NewFileSystem()
	fs.Mkdir("/home/alice/docs")
	fs.CreateFile("/home/alice/docs/cv.pdf", 120)
	fs.CreateFile("/home/alice/docs/notes.txt", 30)
	fs.Mkdir("/home/bob")
	fs.CreateFile("/home/bob/todo.txt", 5)
	ls, _ := fs.Ls("/home")
	size, _ := fs.Size("/home")
	txt, _ := fs.Find("/", "txt")
	fmt.Println("ls /home   =", ls)
	fmt.Println("size /home =", size)
	fmt.Println("find *.txt =", txt)
	fs.Delete("/home/alice")
	ls, _ = fs.Ls("/home")
	size, _ = fs.Size("/")
	fmt.Println("after delete =", ls, "size", size)
}
