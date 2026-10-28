package main

import (
	"fmt"
	"time"
)

func main() {
	var now time.Time
	clock := func() time.Time { return now }
	limiters := []struct {
		name string
		r    RateLimiter
	}{
		{"token-bucket(3, 2/s)", NewTokenBucket(3, 2, clock)},
		{"fixed-window(3/s)", NewFixedWindow(3, time.Second, clock)},
		{"sliding-log(3/s)", NewSlidingWindow(3, time.Second, clock)},
	}
	for _, l := range limiters {
		line := fmt.Sprintf("%-22s", l.name)
		for _, ms := range []int64{0, 100, 200, 300, 600, 1000, 1100, 1250} {
			now = time.UnixMilli(ms)
			mark := "✗"
			if l.r.Allow("user-1") {
				mark = "✓"
			}
			line += fmt.Sprintf("%d%s ", ms, mark)
		}
		fmt.Println(line)
	}
}
