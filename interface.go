package bitvector

import "io"

// Interface describes bit array interface.
type Interface interface {
	io.ReaderFrom
	io.WriterTo
	// Set writes new bit at given position.
	Set(uint64) bool
	// Xor applies xor at given position.
	Xor(uint64) bool
	// Unset clears bit at given position.
	Unset(uint64) bool
	// Get reads bit value from given position.
	Get(uint64) uint8
	// Size returns number of items added to the vector.
	Size() uint64
	// Capacity returns total capacity of the vector.
	Capacity() uint64
	// Popcnt returns population count (number of set bits) in the vector.
	Popcnt() uint64
	// Difference returns count of different bits between two vectors.
	Difference(p Interface) (uint64, error)
	// Merge applies bitwise OR operation with vector p.
	Merge(p Interface) error
	// Clone returns a copy of the bit array.
	Clone() Interface
	// Reset resets the whole bit array.
	Reset()
}
