package bitvector

import "io"

type roaringVector struct{}

func (vec *roaringVector) Set(uint64) bool {
	// todo implement me
	return false
}

func (vec *roaringVector) Xor(uint64) bool {
	// todo implement me
	return false
}

func (vec *roaringVector) Unset(uint64) bool {
	// todo implement me
	return false
}

func (vec *roaringVector) Get(uint64) uint8 {
	// todo implement me
	return 0
}

func (vec *roaringVector) Size() uint64 {
	// todo implement me
	return 0
}

func (vec *roaringVector) Capacity() uint64 {
	// todo implement me
	return 0
}

func (vec *roaringVector) Popcnt() uint64 {
	// todo implement me
	return 0
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
	// todo implement me
	return nil
}

func (vec *roaringVector) Reset() {
	// todo implement me
}

func (vec *roaringVector) ReadFrom(r io.Reader) (n int64, err error) {
	// todo implement me
	return 0, err
}

func (vec *roaringVector) WriteTo(w io.Writer) (n int64, err error) {
	// todo implement me
	return 0, err
}
