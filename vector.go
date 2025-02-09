package bitvector

import (
	"io"

	"github.com/koykov/openrt"
)

// Vector represents simple bit array implementation without race protection. It means you may do concurrent read, but
// cannot do simultaneous read/write operations.
type Vector struct {
	buf []uint8
}

// NewVector make new bit array with given size.
func NewVector(size uint64) (*Vector, error) {
	if size == 0 {
		return nil, ErrZeroSize
	}
	return &Vector{buf: make([]uint8, size/8+1)}, nil
}

// Set writes new bit at given position.
func (vec *Vector) Set(i uint64) bool {
	if len(vec.buf) <= int(i/8) {
		return false
	}
	vec.buf[i/8] |= 1 << (i % 8)
	return true
}

// Unset clears bit at given position.
func (vec *Vector) Unset(i uint64) bool {
	if len(vec.buf) <= int(i/8) {
		return false
	}
	vec.buf[i/8] &^= 1 << (i % 8)
	return true
}

// Get returns bit value from given position.
func (vec *Vector) Get(i uint64) uint8 {
	if len(vec.buf) <= int(i/8) {
		return 0
	}
	return (vec.buf[i/8] & (1 << (i % 8))) >> (i % 8)
}

// Reset resets the whole bit array.
func (vec *Vector) Reset() {
	if len(vec.buf) == 0 {
		return
	}
	openrt.Memclr(vec.buf)
}

func (vec *Vector) ReadFrom(r io.Reader) (int64, error) {
	n, err := r.Read(vec.buf)
	return int64(n), err
}

func (vec *Vector) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(vec.buf)
	return int64(n), err
}
