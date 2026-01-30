package bitvector

import (
	"encoding/binary"
	"io"
	"sort"
	"unsafe"
)

type bitmap struct {
	buf  []uint32
	uniq uint64
}

func (b *bitmap) add(x uint32) {
	const maxSize = 4096
	n := len(b.buf)
	if n == 0 {
		return
	}
	if n < maxSize && b.buf[n-1] < x {
		b.buf = append(b.buf, x)
		return
	}
	pos := sort.Search(n, func(i int) bool { return b.buf[i] == x })
	if pos < 0 {
		if n >= maxSize {
			// todo add new bitmap
			return
		}
		pos1 := -pos - 1
		b.buf = append(b.buf, 0)
		copy(b.buf[pos1+1:], b.buf[pos1:])
		b.buf[pos1] = x
	}
}

func (b *bitmap) remove(x uint32) bool {
	i := sort.Search(len(b.buf), func(i int) bool { return b.buf[i] == x })
	if i < 0 {
		return false
	}
	b.buf = append(b.buf[:i], b.buf[i+1:]...)
	return true
}

func (b *bitmap) index(x uint32) int {
	return sort.Search(len(b.buf), func(i int) bool { return b.buf[i] == x })
}

func (b *bitmap) clone() *bitmap {
	return &bitmap{
		buf:  append([]uint32{}, b.buf...),
		uniq: b.uniq,
	}
}

func (b *bitmap) size() int {
	return len(b.buf)
}

func (b *bitmap) writeTo(w io.Writer) (n int64, err error) {
	var buf [16]byte
	binary.LittleEndian.PutUint64(buf[0:8], b.uniq)
	binary.LittleEndian.PutUint64(buf[8:16], uint64(len(b.buf)))
	var n1 int
	if n1, err = w.Write(buf[:]); err != nil {
		return
	}
	n += int64(n1)

	if len(b.buf) == 0 {
		return
	}

	type h struct {
		p    uintptr
		l, c int
	}
	h1 := *(*h)(unsafe.Pointer(&b.buf))
	h1.l *= 4
	h1.c *= 4
	buf1 := *(*[]byte)(unsafe.Pointer(&h1))
	if n1, err = w.Write(buf1); err != nil {
		return
	}
	n += int64(n1)

	return
}

func (b *bitmap) readFrom(r io.Reader) (n int64, err error) {
	var buf [16]byte
	var n1 int
	n1, err = r.Read(buf[:])
	n += int64(n1)
	if err != nil {
		return
	}

	uniq, ln := binary.LittleEndian.Uint64(buf[0:8]), binary.LittleEndian.Uint64(buf[8:16])
	b.uniq = uniq
	b.buf = make([]uint32, ln)
	type h struct {
		p    uintptr
		l, c int
	}
	h1 := *(*h)(unsafe.Pointer(&b.buf))
	h1.l *= 4
	h1.c *= 4
	buf1 := *(*[]byte)(unsafe.Pointer(&h1))
	n1, err = r.Read(buf1)
	n += int64(n1)

	return
}

func (b *bitmap) reset() {
	b.uniq = 0
	b.buf = b.buf[:0]
}
