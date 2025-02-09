package bitvector

import (
	"io"
	"sync/atomic"
)

const blockSz = 4096

// ConcurrentVector represents concurrent bit array implementation with race protection. Simultaneous read/write
// operations are possible.
type ConcurrentVector struct {
	buf []uint32
	blk [blockSz]byte
	lim uint64
}

// NewConcurrentVector make new concurrent bit array with given size. Param writeAttemptsLimit is the maximum number of
// attempts of atomic writes the bit value.
func NewConcurrentVector(size, writeAttemptsLimit uint64) (*ConcurrentVector, error) {
	if size == 0 {
		return nil, ErrZeroSize
	}
	return &ConcurrentVector{
		buf: make([]uint32, size/32+1),
		lim: writeAttemptsLimit,
	}, nil
}

// Set writes new bit at given position.
func (vec *ConcurrentVector) Set(i uint64) bool {
	if len(vec.buf) <= int(i/32) {
		return false
	}
	var j uint64
	for j = 0; j < vec.lim; j++ {
		o := atomic.LoadUint32(&vec.buf[i/32])
		n := o | 1<<(i%32)
		if atomic.CompareAndSwapUint32(&vec.buf[i/32], o, n) {
			return true
		}
	}
	return false
}

// Unset clears bit at given position.
func (vec *ConcurrentVector) Unset(i uint64) bool {
	if len(vec.buf) <= int(i/32) {
		return false
	}
	var j uint64
	for j = 0; j < vec.lim; j++ {
		o := atomic.LoadUint32(&vec.buf[i/32])
		n := o &^ 1 << (i % 32)
		if atomic.CompareAndSwapUint32(&vec.buf[i/32], o, n) {
			return true
		}
	}
	return false
}

// Get returns bit value from given position.
func (vec *ConcurrentVector) Get(i uint64) uint8 {
	if len(vec.buf) <= int(i/32) {
		return 0
	}
	return uint8((atomic.LoadUint32(&vec.buf[i/32]) & (1 << (i % 32))) >> (i % 32))
}

// Reset resets the whole bit array.
func (vec *ConcurrentVector) Reset() {
	n := len(vec.buf)
	if n == 0 {
		return
	}
	_ = vec.buf[n-1]
	for i := 0; i < n; i++ {
		atomic.StoreUint32(&vec.buf[i], 0)
	}
}

func (vec *ConcurrentVector) ReadFrom(r io.Reader) (int64, error) {
	// todo implement me
	return 0, nil
}

func (vec *ConcurrentVector) WriteTo(w io.Writer) (int64, error) {
	// todo implement me
	return 0, nil
}
