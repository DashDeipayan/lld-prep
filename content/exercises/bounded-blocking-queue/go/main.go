package main

import "fmt"

func main() {
	q := NewBoundedQueue[int](2)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			x, ok := q.Take()
			if !ok {
				fmt.Println("consumer: queue closed and drained")
				return
			}
			fmt.Println("consumed", x)
		}
	}()
	for i := 1; i <= 5; i++ {
		q.Put(i)
		fmt.Println("produced", i)
	}
	q.Close()
	<-done
}
