package bitvector

type Interface interface {
	Set(uint64) bool
	Clear(uint64) bool
	Get(uint64) uint8
	Reset()
}
