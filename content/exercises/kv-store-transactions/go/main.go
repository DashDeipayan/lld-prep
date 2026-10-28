package main

import "fmt"

func main() {
	db := NewKVStore()
	db.Set("a", "10")
	fmt.Println("count 10 =", db.Count("10"))
	db.Begin()
	db.Set("a", "20")
	db.Begin()
	db.Delete("a")
	v, ok := db.Get("a")
	fmt.Println("inner: a =", v, ok)
	db.Rollback()
	v, _ = db.Get("a")
	fmt.Println("after inner rollback: a =", v)
	db.Rollback()
	v, _ = db.Get("a")
	fmt.Println("after outer rollback: a =", v)
	db.Begin()
	db.Set("b", "10")
	db.Commit()
	fmt.Println("count 10 =", db.Count("10"))
	fmt.Println("rollback:", db.Rollback())
}
