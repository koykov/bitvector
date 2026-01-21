package bitvector

import "testing"

func TestBitslice(t *testing.T) {
	t.Run("insert", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			bs := &bitslice{len: 0, buf: make([]uint64, 1)}
			bs.insert(0, true)

			if bs.len != 1 {
				t.Errorf("Expected length 1, got %d", bs.len)
			}
			if (bs.buf[0] & 1) == 0 {
				t.Error("Bit at position 0 should be 1")
			}
		})

		t.Run("at begin", func(t *testing.T) {
			bs := &bitslice{len: 3, buf: []uint64{0b101}}
			bs.insert(0, false)

			if bs.len != 4 {
				t.Errorf("Expected length 4, got %d", bs.len)
			}

			expectedBits := []bool{false, true, false, true}
			for i, expected := range expectedBits {
				bit := (bs.buf[0] >> uint(i)) & 1
				isSet := bit == 1
				if isSet != expected {
					t.Errorf("Bit at position %d: expected %v, got %v", i, expected, isSet)
				}
			}
		})

		t.Run("in middle", func(t *testing.T) {
			bs := &bitslice{len: 4, buf: []uint64{0b1101}}
			bs.insert(2, false)

			if bs.len != 5 {
				t.Errorf("Expected length 5, got %d", bs.len)
			}

			expectedBits := []bool{true, false, false, true, true}
			for i, expected := range expectedBits {
				bit := (bs.buf[0] >> uint(i)) & 1
				isSet := bit == 1
				if isSet != expected {
					t.Errorf("Bit at position %d: expected %v, got %v", i, expected, isSet)
				}
			}
		})

		t.Run("at end", func(t *testing.T) {
			bs := &bitslice{len: 3, buf: []uint64{0b101}}
			bs.insert(3, true)

			if bs.len != 4 {
				t.Errorf("Expected length 4, got %d", bs.len)
			}

			expectedBits := []bool{true, false, true, true}
			for i, expected := range expectedBits {
				bit := (bs.buf[0] >> uint(i)) & 1
				isSet := bit == 1
				if isSet != expected {
					t.Errorf("Bit at position %d: expected %v, got %v", i, expected, isSet)
				}
			}
		})

		t.Run("expand", func(t *testing.T) {
			bs := &bitslice{len: 64, buf: []uint64{0xFFFFFFFFFFFFFFFF}}

			bs.insert(32, false)

			if bs.len != 65 {
				t.Errorf("Expected length 65, got %d", bs.len)
			}
			if len(bs.buf) < 2 {
				t.Errorf("Buffer should have at least 2 words, got %d", len(bs.buf))
			}

			for i := 0; i < 32; i++ {
				wordIdx := i / 64
				bitIdx := uint(i % 64)
				bit := (bs.buf[wordIdx] >> bitIdx) & 1
				if bit != 1 {
					t.Errorf("Bit at position %d should be 1, got 0", i)
				}
			}

			bit32 := (bs.buf[0] >> 32) & 1
			if bit32 != 0 {
				t.Error("Bit at position 32 should be 0")
			}

			for i := 33; i < 65; i++ {
				wordIdx := i / 64
				bitIdx := uint(i % 64)
				bit := (bs.buf[wordIdx] >> bitIdx) & 1
				if bit != 1 {
					t.Errorf("Bit at position %d should be 1, got 0", i)
				}
			}
		})

		t.Run("at bound", func(t *testing.T) {
			bs := &bitslice{len: 64, buf: []uint64{0, 0}}

			bs.buf[0] = 1 << 63
			bs.len = 64

			bs.insert(64, true)

			if bs.len != 65 {
				t.Errorf("Expected length 65, got %d", bs.len)
			}

			if bs.buf[0] != (1 << 63) {
				t.Errorf("First word incorrect: got %b, expected %b", bs.buf[0], uint64(1)<<63)
			}

			if (bs.buf[1] & 1) != 1 {
				t.Error("First bit of second word should be 1")
			}
		})

		t.Run("multiple", func(t *testing.T) {
			bs := &bitslice{len: 0, buf: make([]uint64, 1)}

			bs.insert(0, true)
			bs.insert(0, false)
			bs.insert(2, true)
			bs.insert(1, true)

			if bs.len != 4 {
				t.Errorf("Expected length 4, got %d", bs.len)
			}

			expectedBits := []bool{false, true, true, true}
			for i, expected := range expectedBits {
				bit := (bs.buf[0] >> uint(i)) & 1
				isSet := bit == 1
				if isSet != expected {
					t.Errorf("Bit at position %d: expected %v, got %v", i, expected, isSet)
				}
			}
		})

		t.Run("many bits", func(t *testing.T) {
			bs := &bitslice{len: 0, buf: make([]uint64, 2)}

			for i := 0; i < 128; i++ {
				bs.insert(i, i%2 == 0)
			}

			if bs.len != 128 {
				t.Errorf("Expected length 128, got %d", bs.len)
			}

			for i := 0; i < 128; i++ {
				wordIdx := i / 64
				bitIdx := uint(i % 64)
				bit := (bs.buf[wordIdx] >> bitIdx) & 1
				expected := uint64(0)
				if i%2 == 0 {
					expected = 1
				}
				if bit != expected {
					t.Errorf("Bit at position %d: expected %d, got %d", i, expected, bit)
					break
				}
			}
		})

		t.Run("multi word shift", func(t *testing.T) {
			bs := &bitslice{
				len: 192,
				buf: []uint64{
					0xAAAAAAAAAAAAAAAA,
					0x5555555555555555,
					0xFFFFFFFFFFFFFFFF,
				},
			}

			bs.insert(100, false)

			if bs.len != 193 {
				t.Errorf("Expected length 193, got %d", bs.len)
			}

			testCases := []struct {
				pos      int
				expected uint64
			}{
				{0, 0},
				{1, 1},
				{63, 0},
				{64, 1},
				{99, 0},
				{100, 0},
				{101, 1},
				{191, 1},
				{192, 1},
			}

			for _, tc := range testCases {
				wordIdx := tc.pos / 64
				bitIdx := uint(tc.pos % 64)
				bit := (bs.buf[wordIdx] >> bitIdx) & 1
				if bit != tc.expected {
					t.Errorf("Bit at position %d: expected %d, got %d", tc.pos, tc.expected, bit)
				}
			}
		})
	})
}
