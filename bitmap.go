package bitvector

import "sort"

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
	pos := sort.Search(n, func(i int) bool {
		return b.buf[i] == x
	})
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

func (b *bitmap) index(x uint32) int {
	return sort.Search(len(b.buf), func(i int) bool {
		return b.buf[i] == x
	})
}

func (b *bitmap) clone() *bitmap {
	return &bitmap{
		buf:  append([]uint32{}, b.buf...),
		uniq: b.uniq,
	}
}
