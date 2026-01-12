package bitvector

import (
	"encoding/binary"
	"io"
	"math"
	"math/bits"
	"unsafe"

	"github.com/koykov/simd/bitwise64"
	"github.com/koykov/simd/hamming64"
	"github.com/koykov/simd/memclr64"
	"github.com/koykov/simd/popcnt64"
)

const (
	vectorDumpSignature = 0x65a5cc221b100738
	vectorDumpVersion   = 1.0
)

// vector represents simple bit array implementation without race protection. It means you may do concurrent read, but
// cannot do simultaneous read/write operations.
type vector struct {
	buf  []uint8
	c, s uint64
}

// NewVector make new bit array with given size.
func NewVector(size uint64) (Interface, error) {
	if size == 0 {
		return nil, ErrZeroSize
	}
	return &vector{
		buf: make([]uint8, size/8+1),
		c:   size,
	}, nil
}

// Set writes new bit at given position.
func (vec *vector) Set(i uint64) bool {
	if len(vec.buf) <= int(i/8) {
		return false
	}
	vec.buf[i/8] |= 1 << uint8(i%8)
	vec.s++
	return true
}

// Xor applies xor at given position.
func (vec *vector) Xor(i uint64) bool {
	if len(vec.buf) <= int(i/8) {
		return false
	}
	vec.buf[i/8] ^= 1 << uint8(i%8)
	return true
}

// Unset clears bit at given position.
func (vec *vector) Unset(i uint64) bool {
	if len(vec.buf) <= int(i/8) {
		return false
	}
	vec.buf[i/8] &^= 1 << uint8(i%8)
	vec.s--
	return true
}

// Get returns bit value from given position.
func (vec *vector) Get(i uint64) uint8 {
	if len(vec.buf) <= int(i/8) {
		return 0
	}
	return (vec.buf[i/8] & (1 << (i % 8))) >> (i % 8)
}

// Size returns number of items added to the vector.
func (vec *vector) Size() uint64 {
	return vec.s
}

// Capacity returns total capacity of the vector.
func (vec *vector) Capacity() uint64 {
	return uint64(len(vec.buf)) * 8
}

// Popcnt returns population count (number of set bits) in the vector.
func (vec *vector) Popcnt() (r uint64) {
	buf := vec.buf
	n := len(buf)
	if n == 0 {
		return
	}
	_ = buf[n-1]
	if n > 8 {
		// Interpret bytes slice as uint64 slice.
		type sh struct {
			p    uintptr
			l, c int
		}
		n8 := n / 8
		h := sh{p: uintptr(unsafe.Pointer(&buf[0])), l: n8, c: n8}
		buf64 := *(*[]uint64)(unsafe.Pointer(&h))
		// Apply vectorised population count over uint64 slice.
		r += popcnt64.Count(buf64)
		// Rest of bytes will process below.
		buf = buf[n8*8:]
	}
	for i := 0; i < len(buf); i++ {
		r += uint64(bits.OnesCount8(buf[i]))
	}
	return
}

func (vec *vector) Difference(other Interface) (r uint64, err error) {
	var ovec *vector
	switch x := any(other).(type) {
	case *vector:
		ovec = x
	default:
		err = ErrWrongType
		return
	}
	if vec.c != ovec.c {
		err = ErrNotEqualSize
		return
	}
	buf := vec.buf
	obuf := ovec.buf
	diff := hamming64.DistanceBytes(buf, obuf)
	r = uint64(diff)
	return
}

func (vec *vector) Merge(other Interface) error {
	var ovec *vector
	switch x := any(other).(type) {
	case *vector:
		ovec = x
	default:
		return ErrWrongType
	}
	buf := vec.buf
	obuf := ovec.buf
	bitwise64.OrBytes(buf, obuf)
	return nil
}

func (vec *vector) Clone() Interface {
	clone := &vector{
		buf: make([]uint8, len(vec.buf)),
		c:   vec.c,
		s:   vec.s,
	}
	copy(clone.buf, vec.buf)
	return clone
}

// Reset resets the whole bit array.
func (vec *vector) Reset() {
	if len(vec.buf) == 0 {
		return
	}
	memclr64.ClearBytes(vec.buf)
}

func (vec *vector) ReadFrom(r io.Reader) (n int64, err error) {
	var (
		buf [32]byte
		m   int
	)
	m, err = r.Read(buf[:])
	n += int64(m)
	if err != nil {
		return n, err
	}

	sign, ver, c, s := binary.LittleEndian.Uint64(buf[0:8]), binary.LittleEndian.Uint64(buf[8:16]),
		binary.LittleEndian.Uint64(buf[16:24]), binary.LittleEndian.Uint64(buf[24:32])

	if sign != vectorDumpSignature {
		return n, ErrInvalidSignature
	}
	if ver != math.Float64bits(vectorDumpVersion) {
		return n, ErrVersionMismatch
	}
	vec.c, vec.s = c, s

	if uint64(len(vec.buf)) < c/8+1 {
		vec.buf = make([]uint8, c/8+1)
	}

	m, err = r.Read(vec.buf)
	n += int64(m)
	if err == io.EOF {
		err = nil
	}
	return
}

func (vec *vector) WriteTo(w io.Writer) (n int64, err error) {
	var (
		buf [32]byte
		m   int
	)
	binary.LittleEndian.PutUint64(buf[0:8], vectorDumpSignature)
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(vectorDumpVersion))
	binary.LittleEndian.PutUint64(buf[16:24], vec.c)
	binary.LittleEndian.PutUint64(buf[24:32], vec.s)
	m, err = w.Write(buf[:])
	n += int64(m)
	if err != nil {
		return int64(m), err
	}

	m, err = w.Write(vec.buf)
	n += int64(m)
	return
}
