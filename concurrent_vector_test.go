package bitvector

import (
	"context"
	"math"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
)

func TestConcurrentVector(t *testing.T) {
	prepare := func(size uint) *concurrentVector {
		vec, _ := NewConcurrentVector(10, 0)
		vec.Set(3)
		vec.Set(5)
		vec.Set(7)
		vec.Set(9)
		return any(vec).(*concurrentVector)
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
	t.Run("popcnt", func(t *testing.T) {
		vec := prepare(10)
		if vec.Popcnt() != 4 {
			t.Fail()
		}
	})
	t.Run("difference", func(t *testing.T) {
		vec0 := prepare(128)
		vec1 := prepare(128)
		vec1.Reset()
		if diff, err := vec0.Difference(vec1); diff != 4 || err != nil {
			t.Errorf("difference error: %v, %v", diff, err)
		}
	})
	t.Run("merge", func(t *testing.T) {
		vec0 := prepare(10)
		vec1 := prepare(10)
		vec0.Reset()
		vec0.Set(0)
		vec0.Set(8)

		vec1.Reset()
		vec1.Set(1)
		vec1.Set(9)

		err := vec0.Merge(vec1)
		if err != nil {
			t.Error(err)
		}
		if vec0.Get(1) != 1 {
			t.FailNow()
		}
		if vec0.Get(9) != 1 {
			t.FailNow()
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
		if n != 44 {
			t.Fail()
		}
	})
	t.Run("reader", func(t *testing.T) {
		vec, _ := NewConcurrentVector(10, 0)
		f, err := os.OpenFile("testdata/concurrent_vector.bin", os.O_RDONLY, 0644)
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
	pow := func(x, y int) int {
		return int(math.Pow(float64(x), float64(y)))
	}

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
	b.Run("reset", func(b *testing.B) {
		const base = 1000
		for i := 0; i < 7; i++ {
			sz := base * pow(10, i)
			b.Run(strconv.Itoa(sz), func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(sz))
				vec, _ := NewConcurrentVector(uint64(sz), 0)
				for j := 0; j < b.N; j++ {
					vec.Reset()
				}
			})
		}
	})
	b.Run("parallel io", func(b *testing.B) {
		b.ReportAllocs()

		const size = 1e6
		vec, _ := NewConcurrentVector(size, 3)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func(ctx context.Context, vec Interface) {
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
	b.Run("popcnt", func(b *testing.B) {
		const base = 1000
		for i := 0; i < 7; i++ {
			b.Run(strconv.Itoa(base*pow(10, i)), func(b *testing.B) {
				b.ReportAllocs()
				vec, _ := NewConcurrentVector(uint64(base*pow(10, i)), 0)
				for j := 0; j < b.N; j++ {
					vec.Popcnt()
				}
			})
		}
	})
	b.Run("difference", func(b *testing.B) {
		const base = 1000
		for i := 0; i < 7; i++ {
			sz := base * pow(10, i)
			b.Run(strconv.Itoa(sz), func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(sz))
				vec, _ := NewConcurrentVector(uint64(sz), 0)
				clone := vec.Clone()
				clone.Reset()
				for j := 0; j < b.N; j++ {
					_, _ = vec.Difference(clone)
				}
			})
		}
	})
	b.Run("merge", func(b *testing.B) {
		const base = 1000
		for i := 0; i < 7; i++ {
			sz := base * pow(10, i)
			b.Run(strconv.Itoa(sz), func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(sz))
				vec, _ := NewConcurrentVector(uint64(sz), 0)
				clone := vec.Clone()
				vec.Reset()
				for j := 0; j < b.N; j++ {
					_ = vec.Merge(clone)
				}
			})
		}
	})
}
