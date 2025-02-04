package bitvector

import "github.com/koykov/openrt"

type Vector struct {
	buf []uint8
}

func NewVector(size uint64) (*Vector, error) {
	if size == 0 {
		return nil, ErrZeroSize
	}
	return &Vector{buf: make([]uint8, size/8+1)}, nil
}

func (vec *Vector) Set(i uint64) bool {
	vec.buf[i/8] |= 1 << (i % 8)
	return true
}

func (vec *Vector) Clear(i uint64) bool {
	vec.buf[i/8] &^= 1 << (i % 8)
	return true
}

func (vec *Vector) Get(i uint64) uint8 {
	return (vec.buf[i/8] & (1 << (i % 8))) >> (i % 8)
}

func (vec *Vector) Reset() {
	openrt.Memclr(vec.buf)
}
