package bitvector

import "sync/atomic"

type ConcurrentVector struct {
	buf []uint32
	lim uint64
}

func NewConcurrentVector(size, writeAttemptsLimit uint64) (*ConcurrentVector, error) {
	if size == 0 {
		return nil, ErrZeroSize
	}
	return &ConcurrentVector{
		buf: make([]uint32, size/32+1),
		lim: writeAttemptsLimit,
	}, nil
}

func (vec *ConcurrentVector) Set(i uint64) bool {
	var j uint64
	for j = 0; j < vec.lim; j++ {
		o := atomic.LoadUint32(&vec.buf[j/32])
		n := o | 1<<(i%32)
		if atomic.CompareAndSwapUint32(&vec.buf[j/32], o, n) {
			return true
		}
	}
	return false
}

func (vec *ConcurrentVector) Clear(i uint64) bool {
	var j uint64
	for j = 0; j < vec.lim; j++ {
		o := atomic.LoadUint32(&vec.buf[j/32])
		n := o &^ 1 << (i % 32)
		if atomic.CompareAndSwapUint32(&vec.buf[j/32], o, n) {
			return true
		}
	}
	return false
}

func (vec *ConcurrentVector) Get(i uint64) uint8 {
	return uint8((atomic.LoadUint32(&vec.buf[i/32]) & (1 << (i % 32))) >> (i % 32))
}

func (vec *ConcurrentVector) Reset() {
	n := len(vec.buf)
	_ = vec.buf[n-1]
	for i := 0; i < n; i++ {
		atomic.StoreUint32(&vec.buf[i], 0)
	}
}
