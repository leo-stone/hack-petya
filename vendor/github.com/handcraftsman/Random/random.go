package random

import (
	"crypto/rand"
)

type Random struct {
	nextFeed chan int
}

func NewRandom() *Random {
	r := new(Random)
	bytes := make([]byte, 4)
	r.nextFeed = make(chan int, 100)
	go func() {
		for {
			count, err := rand.Read(bytes)

			if err != nil {
				panic(err)
			}
			if count != 4 {
				continue
			}

			value := 0
			multiplier := 1
			for i := len(bytes) - 1; i >= 0; i-- {
				value += int(bytes[i]) * multiplier
				multiplier *= 256
			}
			r.nextFeed <- value
		}
	}()

	return r
}

func (r *Random) Intn(exclusiveMax int) int {
	value := <-r.nextFeed
	if value < 0 {
		value = -value
	}
	if value >= exclusiveMax {
		value %= exclusiveMax
	}
	return value
}
