package bitvector

// Interface describes bit array interface.
type Interface interface {
	// Set writes new bit at given position.
	Set(uint64) bool
	// Unset clears bit at given position.
	Unset(uint64) bool
	// Get reads bit value from given position.
	Get(uint64) uint8
	// Reset resets the whole bit array.
	Reset()
}
