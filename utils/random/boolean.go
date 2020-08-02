package random

import (
	"math/rand"
	"time"
)

type boolgen struct {
	src rand.Source
	cache int64
	remaining int
}

var bg *boolgen

func init() {
	bg = &boolgen{src: rand.NewSource(time.Now().UnixNano())}
}

// Bool - generates a random boolean
// not safe for concurrent use
func (b *boolgen) gen() bool {
	if b.remaining == 0 {
		b.cache, b.remaining = b.src.Int63(), 63
	}

	result := b.cache&0x01 == 1
	b.cache >>= 1

	return result
}

func Bool() bool {
	return bg.gen()
}

