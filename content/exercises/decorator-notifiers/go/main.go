package main

import (
	"errors"
	"fmt"
	"time"
)

func main() {
	now := time.Unix(0, 0)
	calls := 0
	flakySms := SenderFunc(func(to, msg string) error {
		calls++
		if calls%3 != 0 {
			return errors.New("gateway timeout")
		}
		fmt.Println("  SMS ->", to+":", msg)
		return nil
	})
	var log []string
	s := WithLogging(WithRateLimit(WithRetry(flakySms, 3), func() time.Time { return now }, 2, time.Hour), &log)
	for i := 1; i <= 3; i++ {
		if err := s.Send("+1555", fmt.Sprint("promo ", i)); err != nil {
			fmt.Println("  rejected:", err)
		}
	}
	now = now.Add(time.Hour)
	s.Send("+1555", "promo 4")
	for _, l := range log {
		fmt.Println(l)
	}
}
