package procs

import (
	"math/rand/v2"
	"time"
)

func Delay[T any](seconds int, in <-chan T, out chan<- T) {
	for {
		v := <-in
		if seconds > 0 {
			time.Sleep(time.Duration(rand.IntN(seconds)) * time.Second)
		}
		out <- v
	}
}
