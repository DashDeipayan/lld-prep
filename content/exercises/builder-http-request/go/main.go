package main

import (
	"fmt"
	"time"
)

func main() {
	req, err := NewRequest("https://api.example.com/orders",
		WithMethod("post"),
		WithHeader("Content-Type", "application/json"),
		WithHeader("X-Request-Id", "abc-123"),
		WithTimeout(5*time.Second),
		WithBody(`{"item":"book"}`),
	)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(req.Method(), req.URL())
	fmt.Println("headers =", req.Headers())
	fmt.Println("timeout =", req.Timeout(), "body =", req.Body())
}
