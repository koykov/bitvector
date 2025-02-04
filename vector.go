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
	if len(vec.buf) <= int(i/8) {
		return false
	}
	vec.buf[i/8] |= 1 << (i % 8)
	return true
}

func (vec *Vector) Clear(i uint64) bool {
	if len(vec.buf) <= int(i/8) {
		return false
	}
	vec.buf[i/8] &^= 1 << (i % 8)
	return true
}

func (vec *Vector) Get(i uint64) uint8 {
	if len(vec.buf) <= int(i/8) {
		return 0
	}
	return (vec.buf[i/8] & (1 << (i % 8))) >> (i % 8)
}

func (vec *Vector) Reset() {
	if len(vec.buf) == 0 {
		return
	}
	openrt.Memclr(vec.buf)
}
