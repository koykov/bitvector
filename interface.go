package bitvector

// Interface describes bit array interface.
type Interface interface {
	// Set writes new bit at given position.
	Set(uint64) bool
	// Clear clears bit at given position.
	Clear(uint64) bool
	// Get returns a bit value from given position.
	Get(uint64) uint8
	// Reset resets the whole bit array.
	Reset()
}
