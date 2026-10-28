package main

import "fmt"

func main() {
	today := 100
	lib := NewLibrary(NewPerDayFine(10, 200), func() int { return today },
		NotifierFunc(func(m, msg string) { fmt.Println("  notify", m+":", msg) }))
	lib.AddBook("978-0134685991", "Effective Java", "Joshua Bloch")
	lib.AddBook("978-0201633610", "Design Patterns", "Gamma, Helm, Johnson, Vlissides")
	lib.AddCopy("978-0134685991", "EJ-1")
	lib.AddMember("asha", 3)
	lib.AddMember("ravi", 3)
	loan, _ := lib.Borrow("asha", "978-0134685991")
	fmt.Printf("%+v\n", loan)
	lib.Reserve("ravi", "978-0134685991")
	today += 20
	fine, _ := lib.GiveBack("EJ-1")
	fmt.Println("fine on return =", fine)
	l2, _ := lib.Borrow("ravi", "978-0134685991")
	fmt.Printf("ravi borrows: %+v\n", l2)
	fmt.Println("search 'design' ->", lib.Search("design"))
}
