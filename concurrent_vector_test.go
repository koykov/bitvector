package bitvector

import (
	"context"
	"math"
	"os"
	"sync/atomic"
	"testing"
)

func TestConcurrentVector(t *testing.T) {
	prepare := func(size uint) *ConcurrentVector {
		vec, _ := NewConcurrentVector(10, 0)
		vec.Set(3)
		vec.Set(5)
		vec.Set(7)
		vec.Set(9)
		return vec
	}
	t.Run("set", func(t *testing.T) {
		vec := prepare(10)
		if vec.buf[0] != 680 {
			t.Fail()
		}
	})
	t.Run("unset", func(t *testing.T) {
		vec := prepare(10)
		vec.Unset(5)
		if vec.Get(5) != 0 {
			t.Fail()
		}
	})
	t.Run("get", func(t *testing.T) {
		vec := prepare(10)
		chk := map[int]uint8{3: 1, 5: 1, 7: 1, 9: 1}
		for i := 0; i < 10; i++ {
			if chk[i] != vec.Get(uint64(i)) {
				t.Fail()
			}
		}
	})
	t.Run("writer", func(t *testing.T) {
		vec := prepare(10)
		f, err := os.OpenFile("testdata/concurrent_vector.bin", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return
		}
		n, err := vec.WriteTo(f)
		if err != nil {
			t.Fatal(err)
		}
		if n != 4 {
			t.Fail()
		}
	})
	t.Run("reader", func(t *testing.T) {
		vec, _ := NewConcurrentVector(10, 0)
		f, err := os.Open("testdata/concurrent_vector.bin")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = vec.ReadFrom(f); err != nil {
			t.Fatal(err)
		}
		chk := map[int]uint8{3: 1, 5: 1, 7: 1, 9: 1}
		for i := 0; i < 10; i++ {
			if chk[i] != vec.Get(uint64(i)) {
				t.Fail()
			}
		}
	})
}

func BenchmarkConcurrentVector(b *testing.B) {
	b.Run("set", func(b *testing.B) {
		b.ReportAllocs()
		vec, _ := NewConcurrentVector(10, 0)
		for i := 0; i < b.N; i++ {
			vec.Set(9)
		}
	})
	b.Run("unset", func(b *testing.B) {
		b.ReportAllocs()
		vec, _ := NewConcurrentVector(10, 0)
		vec.Set(5)
		for i := 0; i < b.N; i++ {
			vec.Unset(5)
		}
	})
	b.Run("get", func(b *testing.B) {
		b.ReportAllocs()
		vec, _ := NewConcurrentVector(10, 0)
		vec.Set(5)
		for i := 0; i < b.N; i++ {
			vec.Get(5)
		}
	})
	b.Run("parallel io", func(b *testing.B) {
		b.ReportAllocs()

		const size = 1e6
		vec, _ := NewConcurrentVector(size, 3)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func(ctx context.Context, vec *ConcurrentVector) {
			for i := uint64(0); ; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					vec.Set(i % size)
					vec.Unset(i % size)
					vec.Set(i % size)
				}
			}
		}(ctx, vec)

		var i uint64 = math.MaxUint64
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				vec.Get(atomic.AddUint64(&i, 1))
			}
		})
	})
}
