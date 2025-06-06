package bitvector

import (
	"encoding/binary"
	"io"
	"math"
	"math/bits"
	"sync/atomic"
)

const (
	cnVectorDumpSignature = 0xe1aa38d7f1fe3cd9
	cnVectorDumpVersion   = 1.0

	blockSz = 4096
)

// concurrentVector represents concurrent bit array implementation with race protection. Simultaneous read/write
// operations are possible.
type concurrentVector struct {
	buf  []uint32
	blk  [blockSz]byte
	lim  uint64
	c, s uint64
}

// NewConcurrentVector make new concurrent bit array with given size. Param writeAttemptsLimit is the maximum number of
// attempts of atomic writes the bit value.
func NewConcurrentVector(size, writeAttemptsLimit uint64) (Interface, error) {
	if size == 0 {
		return nil, ErrZeroSize
	}
	return &concurrentVector{
		buf: make([]uint32, size/32+1),
		lim: writeAttemptsLimit + 1,
		c:   size,
	}, nil
}

// Set writes new bit at given position.
func (vec *concurrentVector) Set(i uint64) bool {
	if len(vec.buf) <= int(i/32) {
		return false
	}
	for j := uint64(0); j < vec.lim; j++ {
		o := atomic.LoadUint32(&vec.buf[i/32])
		n := o | 1<<(i%32)
		if atomic.CompareAndSwapUint32(&vec.buf[i/32], o, n) {
			atomic.AddUint64(&vec.s, 1)
			return true
		}
	}
	return false
}

// Xor applies xor at given position.
func (vec *concurrentVector) Xor(i uint64) bool {
	if len(vec.buf) <= int(i/32) {
		return false
	}
	for j := uint64(0); j < vec.lim; j++ {
		o := atomic.LoadUint32(&vec.buf[i/32])
		n := o ^ 1<<(i%32)
		if atomic.CompareAndSwapUint32(&vec.buf[i/32], o, n) {
			atomic.AddUint64(&vec.s, 1)
			return true
		}
	}
	return false
}

// Unset clears bit at given position.
func (vec *concurrentVector) Unset(i uint64) bool {
	if len(vec.buf) <= int(i/32) {
		return false
	}
	for j := uint64(0); j < vec.lim; j++ {
		o := atomic.LoadUint32(&vec.buf[i/32])
		n := o &^ 1 << (i % 32)
		if atomic.CompareAndSwapUint32(&vec.buf[i/32], o, n) {
			atomic.AddUint64(&vec.s, math.MaxUint64)
			return true
		}
	}
	return false
}

// Get returns bit value from given position.
func (vec *concurrentVector) Get(i uint64) uint8 {
	if len(vec.buf) <= int(i/32) {
		return 0
	}
	return uint8((atomic.LoadUint32(&vec.buf[i/32]) & (1 << (i % 32))) >> (i % 32))
}

// Size returns number of items added to the vector.
func (vec *concurrentVector) Size() uint64 {
	return atomic.LoadUint64(&vec.s)
}

// Capacity returns total capacity of the vector.
func (vec *concurrentVector) Capacity() uint64 {
	return uint64(len(vec.buf)) * 32
}

// Popcnt returns population count (number of set bits) in the vector.
func (vec *concurrentVector) Popcnt() (r uint64) {
	n := len(vec.buf)
	if n == 0 {
		return
	}
	_ = vec.buf[n-1]
	for i := 0; i < n; i++ {
		v := atomic.LoadUint32(&vec.buf[i])
		r += uint64(bits.OnesCount32(v))
	}
	return
}

func (vec *concurrentVector) Difference(other Interface) (r uint64, err error) {
	var ovec *concurrentVector
	switch x := any(other).(type) {
	case *concurrentVector:
		ovec = x
	default:
		err = ErrWrongType
		return
	}
	if vec.c != ovec.c {
		err = ErrNotEqualSize
		return
	}
	n := len(vec.buf)
	if n == 0 {
		n = len(ovec.buf)
	}
	_ = vec.buf[n-1]
	_ = ovec.buf[n-1]
	for i := 0; i < n; i++ {
		v := atomic.LoadUint32(&vec.buf[i]) ^ atomic.LoadUint32(&ovec.buf[i])
		r += uint64(bits.OnesCount32(v))
	}
	return
}

func (vec *concurrentVector) Clone() Interface {
	clone := &concurrentVector{
		buf: make([]uint32, len(vec.buf)),
		c:   vec.c,
		s:   atomic.LoadUint64(&vec.s),
		lim: vec.lim,
	}
	for i := 0; i < len(vec.buf); i++ {
		atomic.StoreUint32(&clone.buf[i], atomic.LoadUint32(&vec.buf[i]))
	}
	return clone
}

// Reset resets the whole bit array.
func (vec *concurrentVector) Reset() {
	n := len(vec.buf)
	if n == 0 {
		return
	}
	_ = vec.buf[n-1]
	for i := 0; i < n; i++ {
		atomic.StoreUint32(&vec.buf[i], 0)
	}
}

func (vec *concurrentVector) ReadFrom(r io.Reader) (n int64, err error) {
	var (
		buf [40]byte
		m   int
	)
	m, err = r.Read(buf[:])
	n += int64(m)
	if err != nil {
		return n, err
	}

	sign, ver, c, s, lim := binary.LittleEndian.Uint64(buf[0:8]), binary.LittleEndian.Uint64(buf[8:16]),
		binary.LittleEndian.Uint64(buf[16:24]), binary.LittleEndian.Uint64(buf[24:32]),
		binary.LittleEndian.Uint64(buf[32:40])

	if sign != cnVectorDumpSignature {
		return n, ErrInvalidSignature
	}
	if ver != math.Float64bits(cnVectorDumpVersion) {
		return n, ErrVersionMismatch
	}
	vec.c, vec.lim = c, lim
	atomic.StoreUint64(&vec.s, s)

	if cp := c/32 + 1; uint64(len(vec.buf)) < cp {
		vec.buf = make([]uint32, cp)
	}

	for i := 0; ; i += blockSz {
		m, err = r.Read(vec.blk[:])
		n += int64(m)
		if err != nil && err != io.EOF {
			return n, err
		}
		for j := 0; j < m; j += 4 {
			v := binary.LittleEndian.Uint32(vec.blk[j:])
			atomic.StoreUint32(&vec.buf[(i+j)/4], v)
		}
		if err == io.EOF {
			err = nil
			break
		}
	}
	return
}

func (vec *concurrentVector) WriteTo(w io.Writer) (n int64, err error) {
	var (
		buf [40]byte
		m   int
	)
	binary.LittleEndian.PutUint64(buf[0:8], cnVectorDumpSignature)
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(cnVectorDumpVersion))
	binary.LittleEndian.PutUint64(buf[16:24], vec.c)
	binary.LittleEndian.PutUint64(buf[24:32], atomic.LoadUint64(&vec.s))
	binary.LittleEndian.PutUint64(buf[32:40], vec.lim)
	if m, err = w.Write(buf[:]); err != nil {
		return int64(m), err
	}
	n += int64(m)

	var off int
	for i := 0; i < len(vec.buf); i++ {
		v := atomic.LoadUint32(&vec.buf[i])
		binary.LittleEndian.PutUint32(vec.blk[off:], v)
		if off += 4; off == blockSz {
			m, err = w.Write(vec.blk[:off])
			n += int64(m)
			if err != nil {
				return
			}
			if m < blockSz {
				err = io.ErrShortWrite
				return
			}
			off = 0
		}
	}
	if off > 0 {
		m, err = w.Write(vec.blk[:off])
		n += int64(m)
	}
	return
}
