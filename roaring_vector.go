package bitvector

import (
	"encoding/binary"
	"io"
	"math"
	"sort"
	"unsafe"
)

const (
	roaringVectorDumpSignature = 0x9cf814f5923ac3bf
	roaringVectorDumpVersion   = 1.0
)

type roaringVector struct {
	rvector
	cpy rvector
}

type rvector struct {
	keys []uint32
	buf  []*bitmap
	cow  bitslice
}

func (vec *roaringVector) Set(x uint64) bool {
	hib, lob := vec.hibits(x), vec.lobits(x)
	return vec.setHL(hib, lob)
}

func (vec *roaringVector) Xor(uint64) bool {
	return false // not implemented
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
	return vec.getHL(hib, lob)
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
	inst, ok := any(p).(*roaringVector)
	if !ok {
		return 0, ErrWrongType
	}
	var c uint64
	for i := 0; i < len(vec.keys); i++ {
		bi0, bi1 := vec.indexhb(vec.keys[i]), inst.indexhb(vec.keys[i])
		if bi0 < 0 {
			continue
		}
		if bi1 < 0 {
			c += uint64(vec.buf[bi0].size())
		} else {
			b0, b1 := vec.buf[bi0], inst.buf[bi1]
			var j0, j1 int
			for j0 < len(b0.buf) && j1 < len(b1.buf) {
				switch {
				case b0.buf[j0] == b1.buf[j1]:
					j0++
					j1++
				case b0.buf[j0] < b1.buf[j1]:
					c++
					j0++
				case b0.buf[j0] > b1.buf[j1]:
					c++
					j1++
				}
			}
			c += uint64((len(b0.buf) - j0) + (len(b1.buf) - j1))
		}
	}
	for i := 0; i < len(inst.keys); i++ {
		bi0, bi1 := vec.indexhb(inst.keys[i]), inst.indexhb(inst.keys[i])
		if bi1 < 0 {
			continue
		}
		if bi0 < 0 {
			c += uint64(inst.buf[bi1].size())
		}
	}
	return c, nil
}

func (vec *roaringVector) Merge(p Interface) error {
	inst, ok := any(p).(*roaringVector)
	if !ok {
		return ErrWrongType
	}
	vec.copyTo(&vec.cpy)
	for i := 0; i < len(inst.keys); i++ {
		key := inst.keys[i]
		bi := inst.indexhb(key)
		if bi < 0 {
			continue
		}
		b := inst.buf[bi]
		for j := 0; j < len(b.buf); j++ {
			vec.cpy.setHL(key, b.buf[j])
		}
	}
	return nil
}

func (vec *roaringVector) Filter(p Interface) error {
	inst, ok := any(p).(*roaringVector)
	if !ok {
		return ErrWrongType
	}
	vec.cpy.Reset()
	var i0, i1 int
	for i0 < len(vec.keys) && i1 < len(inst.keys) {
		key0, key1 := vec.keys[i0], inst.keys[i1]
		if key0 == key1 {
			bi0, bi1 := vec.indexhb(key0), inst.indexhb(key1)
			if bi0 == -1 || bi1 == -1 {
				continue
			}
			b0, b1 := vec.buf[bi0], inst.buf[bi1]
			var j0, j1 int
			for j0 < len(b0.buf) && j1 < len(b1.buf) {
				switch {
				case b0.buf[j0] == b1.buf[j1]:
					vec.cpy.setHL(key0, b0.buf[j0])
					j0++
					j1++
				case b0.buf[j0] < b1.buf[j1]:
					j0++
				case b0.buf[j0] > b1.buf[j1]:
					j1++
				}
			}
		}
	}
	return nil
}

func (vec *roaringVector) Invert() {
	// can't be implemented
}

func (vec *roaringVector) Clone() Interface {
	cpy := &roaringVector{
		rvector: rvector{
			keys: append([]uint32{}, vec.keys...),
			buf:  make([]*bitmap, len(vec.buf)),
			cow:  vec.cow.clone(),
		},
	}
	for i := 0; i < len(vec.buf); i++ {
		cpy.buf[i] = vec.buf[i].clone()
	}
	return cpy
}

func (vec *roaringVector) ReadFrom(r io.Reader) (n int64, err error) {
	var (
		buf [24]byte
		m   int
	)
	m, err = r.Read(buf[:])
	n += int64(m)
	if err != nil {
		return n, err
	}

	sign, ver, ln := binary.LittleEndian.Uint64(buf[0:8]), binary.LittleEndian.Uint64(buf[8:16]),
		binary.LittleEndian.Uint64(buf[16:24])

	if sign != roaringVectorDumpSignature {
		return n, ErrInvalidSignature
	}
	if ver != math.Float64bits(roaringVectorDumpVersion) {
		return n, ErrVersionMismatch
	}

	vec.keys = make([]uint32, ln)
	h1 := *(*hslice)(unsafe.Pointer(&vec.keys))
	h1.l *= 4
	h1.c *= 4
	buf1 := *(*[]byte)(unsafe.Pointer(&h1))
	m, err = r.Read(buf1)
	n += int64(m)
	if err != nil {
		return
	}

	m, err = r.Read(buf[:8])
	ln = binary.LittleEndian.Uint64(buf[0:8])
	vec.buf = make([]*bitmap, ln)
	var n1 int64
	for i := uint64(0); i < ln; i++ {
		vec.buf[i] = &bitmap{}
		n1, err = vec.buf[i].readFrom(r)
		n += n1
		if err != nil {
			return
		}
	}

	n1, err = vec.cow.readFrom(r)
	n += n1

	return
}

func (vec *roaringVector) WriteTo(w io.Writer) (n int64, err error) {
	var (
		buf [24]byte
		m   int
	)
	binary.LittleEndian.PutUint64(buf[0:8], roaringVectorDumpSignature)
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(roaringVectorDumpVersion))
	binary.LittleEndian.PutUint64(buf[16:24], uint64(len(vec.keys)))
	m, err = w.Write(buf[:])
	n += int64(m)
	if err != nil {
		return int64(m), err
	}

	h1 := *(*hslice)(unsafe.Pointer(&vec.keys))
	h1.l *= 4
	h1.c *= 4
	buf1 := *(*[]byte)(unsafe.Pointer(&h1))
	if m, err = w.Write(buf1); err != nil {
		return
	}
	n += int64(m)

	binary.LittleEndian.PutUint64(buf[0:8], uint64(len(vec.buf)))
	m, err = w.Write(buf[:8])
	n += int64(m)
	if err != nil {
		return int64(m), err
	}
	for i := 0; i < len(vec.buf); i++ {
		n1, err := vec.buf[i].writeTo(w)
		if err != nil {
			return
		}
		n += n1
	}

	n1, err := vec.cow.writeTo(w)
	if err != nil {
		return
	}
	n += n1

	return
}

func (vec *roaringVector) Reset() {
	vec.rvector.Reset()
}

func (vec *rvector) Reset() {
	vec.keys = vec.keys[:0]
	vec.buf = vec.buf[:0]
	vec.cow.reset()
}

func (vec *rvector) hibits(x uint64) uint32 {
	return uint32(x >> 32)
}

func (vec *rvector) lobits(x uint64) uint32 {
	return uint32(x & math.MaxUint32)
}

func (vec *rvector) setHL(hib, lob uint32) bool {
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

func (vec *rvector) getHL(hib, lob uint32) uint8 {
	i := vec.indexhb(hib)
	if i < 0 || i >= len(vec.buf) {
		return 0
	}
	if j := vec.buf[i].index(lob); j >= 0 {
		return 1
	}
	return 0
}

func (vec *rvector) indexhb(hb uint32) int {
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

func (vec *rvector) addhb(i int, hb uint32, bm *bitmap) {
	vec.keys = append(vec.keys, 0)
	copy(vec.keys[i+1:], vec.keys[i:])
	vec.keys[i] = hb

	vec.buf = append(vec.buf, nil)
	copy(vec.buf[i+1:], vec.buf[i:])
	vec.buf[i] = bm

	vec.cow.insert(i, false)
}

func (vec *rvector) copyTo(o *rvector) {
	o.keys = append(o.keys[:0], vec.keys...)
	o.buf = o.buf[:0]
	for i := 0; i < len(vec.buf); i++ {
		o.buf = append(o.buf, vec.buf[i].clone())
	}
	vec.cow.copyTo(&o.cow)
}
