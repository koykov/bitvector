# Bitvector

The package provides implementations of bit arrays of arbitrary length. Bit array allows to manipulate with single bits:
* set bit to 1 in given position
* clear bit to 0 in given position
* get bit value in given position
* and reset the whole array

Currently supports two types of arrays:
* [Vector](vector.go)
* [ConcurrentVector](concurrent_vector.go)

Both have the same [interface](interface.go), but concurrent version supports simultaneously read and write.

## Usage

[Vector](vector.go) example:
```go
import "github.com/koykov/bitvector"

vec := bitvector.NewVector(100)
vec.Set(50)   // set bit value to 1 at position 50
vec.Get(50)   // read bit value at position 50
vec.Clear(50) // clear bit at position 50
```
This type allows simultaneous read and write, but doesn't provide data race protection. A good use case is preliminary
set required bits values and then read them in concurrent manner.

[ConcurrentVector](concurrent_vector.go) example:
```go
import "github.com/koykov/bitvector"

v := bitvector.NewConcurrentVector(100)
go func(){ for { v.Set(50) } }
go func(){ for { v.Get(50) } }
go func(){ for { v.Clear(50) } }
```
In opposite to [Vector](vector.go), this type supports simultaneous read and write and provides data race protection.
It uses atomics inside, thus works without exclusive locks and works fast (check benchmarks).
