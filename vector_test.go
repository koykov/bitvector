package bitvector

import "testing"

func TestVector(t *testing.T) {
	prepare := func(size uint) *Vector {
		vec, _ := NewVector(10)
		vec.Set(3)
		vec.Set(5)
		vec.Set(7)
		vec.Set(9)
		return vec
	}
	t.Run("set", func(t *testing.T) {
		vec := prepare(10)
		if vec.buf[0] != 168 || vec.buf[1] != 2 {
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
		vec.Unset(5)
		if vec.Get(5) != 0 {
			t.Fail()
		}
	})
}

func BenchmarkVector(b *testing.B) {
	b.Run("set", func(b *testing.B) {
		b.ReportAllocs()
		vec, _ := NewVector(10)
		for i := 0; i < b.N; i++ {
			vec.Set(9)
		}
	})
	b.Run("get", func(b *testing.B) {
		b.ReportAllocs()
		vec, _ := NewVector(10)
		vec.Set(5)
		for i := 0; i < b.N; i++ {
			vec.Get(5)
		}
	})
	b.Run("clear", func(b *testing.B) {
		b.ReportAllocs()
		vec, _ := NewVector(10)
		vec.Set(5)
		for i := 0; i < b.N; i++ {
			vec.Unset(5)
		}
	})
}
