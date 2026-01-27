package bitvector

import (
	"io"
	"math"
	"sort"
)

type roaringVector struct {
	keys []uint32
	buf  []*bitmap
	cow  bitslice
}

func NewRoaringVector(size uint64) (Interface, error) {
	if size == 0 {
		return nil, ErrZeroSize
	}
	return &roaringVector{
		// ...
	}, nil
}

func (vec *roaringVector) Set(x uint64) bool {
	hib, lob := vec.hibits(x), vec.lobits(x)
	i := vec.indexhb(hib)
	if i < 0 {
		bm := bitmap{}
		bm.add(lob)
		vec.addhb(-i-1, hib, &bm)
		return true
	}

	var bm *bitmap
	if vec.cow.get(i) {
		bm = vec.buf[i].clone()
	} else {
		bm = vec.buf[i]
	}
	bm.add(lob)
	vec.buf[i] = bm

	return true
}

func (vec *roaringVector) addhb(i int, hb uint32, bm *bitmap) {
	vec.keys = append(vec.keys, 0)
	copy(vec.keys[i+1:], vec.keys[i:])
	vec.keys[i] = hb

	vec.buf = append(vec.buf, nil)
	copy(vec.buf[i+1:], vec.buf[i:])
	vec.buf[i] = bm

	vec.cow.insert(i, false)
}

func (vec *roaringVector) Xor(uint64) bool {
	// todo implement me
	return false
}

func (vec *roaringVector) Unset(x uint64) bool {
	hib, lob := vec.hibits(x), vec.lobits(x)
	i := vec.indexhb(hib)
	if i < 0 {
		return false
	}
	var bm *bitmap
	bm = vec.buf[i]
	if !bm.remove(lob) {
		return false
	}
	if bm.size() == 0 {
		copy(vec.keys[i:], vec.keys[i+1:])
		vec.keys = vec.keys[len(vec.keys)-1:]
		copy(vec.buf[i:], vec.buf[i+1:])
		vec.buf = vec.buf[len(vec.buf)-1:]
		vec.cow.delete(i)
	}
	return false
}

func (vec *roaringVector) Get(x uint64) uint8 {
	hib, lob := vec.hibits(x), vec.lobits(x)
	i := vec.indexhb(hib)
	if i < 0 || i >= len(vec.buf) {
		return 0
	}
	if j := vec.buf[i].index(lob); j >= 0 {
		return 1
	}
	return 0
}

func (vec *roaringVector) Size() uint64 {
	return uint64(len(vec.keys))
}

func (vec *roaringVector) Capacity() uint64 {
	return uint64(cap(vec.keys))
}

func (vec *roaringVector) Popcnt() (c uint64) {
	for i := range vec.buf {
		c += uint64(vec.buf[i].size())
	}
	return
}

func (vec *roaringVector) Difference(p Interface) (uint64, error) {
	// todo implement me
	return 0, nil
}

func (vec *roaringVector) Merge(p Interface) error {
	// todo implement me
	return nil
}

func (vec *roaringVector) Filter(p Interface) error {
	// todo implement me
	return nil
}

func (vec *roaringVector) Invert() {
	// todo implement me
}

func (vec *roaringVector) Clone() Interface {
	cpy := &roaringVector{
		keys: append([]uint32{}, vec.keys...),
		buf:  make([]*bitmap, len(vec.buf)),
		cow:  vec.cow.clone(),
	}
	for i := 0; i < len(vec.buf); i++ {
		cpy.buf[i] = vec.buf[i].clone()
	}
	return cpy
}

func (vec *roaringVector) ReadFrom(r io.Reader) (n int64, err error) {
	// todo implement me
	return 0, err
}

func (vec *roaringVector) WriteTo(w io.Writer) (n int64, err error) {
	// todo implement me
	return 0, err
}

func (vec *roaringVector) Reset() {
	// todo implement me
}

func (vec *roaringVector) hibits(x uint64) uint32 {
	return uint32(x >> 32)
}

func (vec *roaringVector) lobits(x uint64) uint32 {
	return uint32(x & math.MaxUint32)
}

func (vec *roaringVector) indexhb(hb uint32) int {
	n := len(vec.keys)
	if n == 0 {
		return -1
	}
	if hb == vec.keys[n-1] {
		return n - 1
	}
	return sort.Search(n, func(i int) bool {
		return vec.keys[i] == hb
	})
}
