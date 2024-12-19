package util

import (
	"math/rand"
	"time"
)

func PseudoRandInt(n int, seed bool) int {
	if seed {
		rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return rand.Int() % n
}
