package bitvector

import (
	"encoding/binary"
	"io"
	"math"
	"math/bits"
	"unsafe"

	"github.com/koykov/openrt"
	"github.com/koykov/simd/popcnt64"
)

const (
	vectorDumpSignature = 0x65a5cc221b100738
	vectorDumpVersion   = 1.0
)

// Vector represents simple bit array implementation without race protection. It means you may do concurrent read, but
// cannot do simultaneous read/write operations.
type Vector struct {
	buf  []uint8
	c, s uint64
}

// NewVector make new bit array with given size.
func NewVector(size uint64) (*Vector, error) {
	if size == 0 {
		return nil, ErrZeroSize
	}
	return &Vector{
		buf: make([]uint8, size/8+1),
		c:   size,
	}, nil
}

// Set writes new bit at given position.
func (vec *Vector) Set(i uint64) bool {
	if len(vec.buf) <= int(i/8) {
		return false
	}
	vec.buf[i/8] |= 1 << uint8(i%8)
	vec.s++
	return true
}

// Unset clears bit at given position.
func (vec *Vector) Unset(i uint64) bool {
	if len(vec.buf) <= int(i/8) {
		return false
	}
	vec.buf[i/8] &^= 1 << uint8(i%8)
	vec.s--
	return true
}

// Get returns bit value from given position.
func (vec *Vector) Get(i uint64) uint8 {
	if len(vec.buf) <= int(i/8) {
		return 0
	}
	return (vec.buf[i/8] & (1 << (i % 8))) >> (i % 8)
}

// Size returns number of items added to the vector.
func (vec *Vector) Size() uint64 {
	return vec.s
}

// Capacity returns total capacity of the vector.
func (vec *Vector) Capacity() uint64 {
	return uint64(len(vec.buf)) * 8
}

// OnesCount returns number of ones in the vector.
func (vec *Vector) OnesCount() (r uint64) {
	buf := vec.buf
	n := len(buf)
	if n == 0 {
		return
	}
	_ = buf[n-1]
	if n > 8 {
		type sh struct {
			p    uintptr
			l, c int
		}
		n8 := n / 8
		h := sh{p: uintptr(unsafe.Pointer(&buf[0])), l: n8, c: n8}
		buf64 := *(*[]uint64)(unsafe.Pointer(&h))
		r += popcnt64.Count(buf64)
		buf = buf[n8*8:]
	}
	for i := 0; i < len(buf); i++ {
		r += uint64(bits.OnesCount8(buf[i]))
	}
	return
}

// Reset resets the whole bit array.
func (vec *Vector) Reset() {
	if len(vec.buf) == 0 {
		return
	}
	openrt.Memclr(vec.buf)
}

func (vec *Vector) ReadFrom(r io.Reader) (n int64, err error) {
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

func (vec *Vector) WriteTo(w io.Writer) (n int64, err error) {
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
