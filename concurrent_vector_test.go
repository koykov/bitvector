package bitvector

import (
	"context"
	"math"
	"sync/atomic"
	"testing"
)

func TestConcurrentVector(t *testing.T) {
	prepare := func(size uint) *ConcurrentVector {
		vec, _ := NewConcurrentVector(10, 1)
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
	t.Run("get", func(t *testing.T) {
		vec := prepare(10)
		chk := map[int]uint8{3: 1, 5: 1, 7: 1, 9: 1}
		for i := 0; i < 10; i++ {
			if chk[i] != vec.Get(uint64(i)) {
				t.Fail()
			}
		}
	})
	t.Run("clear", func(t *testing.T) {
		vec := prepare(10)
		vec.Clear(5)
		if vec.Get(5) != 0 {
			t.Fail()
		}
	})
}

func BenchmarkConcurrentVector(b *testing.B) {
	b.Run("set", func(b *testing.B) {
		b.ReportAllocs()
		vec, _ := NewConcurrentVector(10, 1)
		for i := 0; i < b.N; i++ {
			vec.Set(9)
		}
	})
	b.Run("get", func(b *testing.B) {
		b.ReportAllocs()
		vec, _ := NewConcurrentVector(10, 1)
		vec.Set(5)
		for i := 0; i < b.N; i++ {
			vec.Get(5)
		}
	})
	b.Run("clear", func(b *testing.B) {
		b.ReportAllocs()
		vec, _ := NewConcurrentVector(10, 1)
		vec.Set(5)
		for i := 0; i < b.N; i++ {
			vec.Clear(5)
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
					vec.Set(i)
					vec.Clear(i)
					vec.Set(i)
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
