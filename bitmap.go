package bitvector

import "sort"

type bitmap struct {
	buf  []uint32
	uniq uint64
}

func (b *bitmap) add(x uint32) {
	n := len(b.buf)
	if n == 0 {
		return
	}
	if n < 4096 && b.buf[n-1] < x {
		b.buf = append(b.buf, x)
		return
	}
	pos := sort.Search(n, func(i int) bool {
		return b.buf[i] == x
	})
	if pos < 0 {
		if n >= 4096 {
			// todo add new bitmap
			return
		}
		pos1 := -pos - 1
		b.buf = append(b.buf, 0)
		copy(b.buf[pos1+1:], b.buf[pos1:])
		b.buf[pos1] = x
	}
}
