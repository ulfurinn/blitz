package blizzard

import (
	"io"
	"math/rand"
	"time"
)

// Randbo creates a stream of non-crypto quality random bytes
type randbo struct {
	rand.Source
}

// New creates a new random reader with a time source.
func NewRand() io.Reader {
	return NewRandFrom(rand.NewSource(time.Now().UnixNano()))
}

// NewFrom creates a new reader from your own rand.Source
func NewRandFrom(src rand.Source) io.Reader {
	return &randbo{src}
}

// Read satisfies io.Reader
func (r *randbo) Read(p []byte) (n int, err error) {
	todo := len(p)
	offset := 0
	for {
		val := int64(r.Int63())
		for i := 0; i < 8; i++ {
			p[offset] = 65 + byte(val)%26
			todo--
			if todo == 0 {
				return len(p), nil
			}
			offset++
			val >>= 8
		}
	}
}
