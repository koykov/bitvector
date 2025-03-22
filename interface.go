package bitvector

import "io"

// Interface describes bit array interface.
type Interface interface {
	io.ReaderFrom
	io.WriterTo
	// Set writes new bit at given position.
	Set(uint64) bool
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
	// Reset resets the whole bit array.
	Reset()
}
