package bitvector

import (
	"encoding/binary"
	"io"
	"math"
	"sync/atomic"
)

const blockSz = 4096

// ConcurrentVector represents concurrent bit array implementation with race protection. Simultaneous read/write
// operations are possible.
type ConcurrentVector struct {
	buf []uint32
	blk [blockSz]byte
	lim uint64
	c   uint64
}

// NewConcurrentVector make new concurrent bit array with given size. Param writeAttemptsLimit is the maximum number of
// attempts of atomic writes the bit value.
func NewConcurrentVector(size, writeAttemptsLimit uint64) (*ConcurrentVector, error) {
	if size == 0 {
		return nil, ErrZeroSize
	}
	return &ConcurrentVector{
		buf: make([]uint32, size/32+1),
		lim: writeAttemptsLimit + 1,
	}, nil
}

// Set writes new bit at given position.
func (vec *ConcurrentVector) Set(i uint64) bool {
	if len(vec.buf) <= int(i/32) {
		return false
	}
	for j := uint64(0); j < vec.lim; j++ {
		o := atomic.LoadUint32(&vec.buf[i/32])
		n := o | 1<<(i%32)
		if atomic.CompareAndSwapUint32(&vec.buf[i/32], o, n) {
			atomic.AddUint64(&vec.c, 1)
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
	for j := uint64(0); j < vec.lim; j++ {
		o := atomic.LoadUint32(&vec.buf[i/32])
		n := o &^ 1 << (i % 32)
		if atomic.CompareAndSwapUint32(&vec.buf[i/32], o, n) {
			atomic.AddUint64(&vec.c, math.MaxUint64)
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

// Size returns number of items added to the vector.
func (vec *ConcurrentVector) Size() uint64 {
	return atomic.LoadUint64(&vec.c)
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
	var n int
	for {
		n1, err := r.Read(vec.blk[:])
		if err != nil && err != io.EOF {
			return int64(n), err
		}
		n += n1
		for i := 0; i < n1; i += 4 {
			v := binary.LittleEndian.Uint32(vec.blk[i:])
			atomic.StoreUint32(&vec.buf[i/4], v)
		}
		if err == io.EOF {
			break
		}
	}
	return int64(n), nil
}

func (vec *ConcurrentVector) WriteTo(w io.Writer) (int64, error) {
	var off, n int
	for i := 0; i < len(vec.buf); i++ {
		v := atomic.LoadUint32(&vec.buf[i])
		binary.LittleEndian.PutUint32(vec.blk[off:], v)
		if off += 4; off == blockSz {
			n1, err := w.Write(vec.blk[:off])
			n += n1
			if err != nil {
				return int64(n), err
			}
			if n1 < blockSz {
				return int64(n), io.ErrShortWrite
			}
			off = 0
		}
	}
	if off > 0 {
		n1, err := w.Write(vec.blk[:off])
		n += n1
		if err != nil {
			return int64(n), err
		}
	}
	return int64(n), nil
}
