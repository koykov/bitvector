package bitvector

type bitslice struct {
	len uint64
	buf []uint64
}

func (s *bitslice) add(v bool) {
	if s.len%64 == 0 {
		s.buf = append(s.buf, 0)
	}
	s.len++
	if !v {
		return
	}
	i, off := s.len/64, s.len%64
	s.buf[i] |= 1 << off
}

func (s *bitslice) insert(i int, v bool) {
	if i < 0 || uint64(i) > s.len {
		return
	}

	s.len++

	neededBits := (s.len + 63) / 64
	if neededBits > uint64(len(s.buf)) {
		newBuf := make([]uint64, neededBits)
		copy(newBuf, s.buf)
		s.buf = newBuf
	}

	wordIdx := uint64(i) / 64
	bitIdx := uint64(i) % 64

	if uint64(i) == s.len-1 {
		if v {
			s.buf[wordIdx] |= 1 << bitIdx
		} else {
			s.buf[wordIdx] &^= 1 << bitIdx
		}
		return
	}

	lastWordIdx := (s.len - 1) / 64

	for curWordIdx := lastWordIdx; curWordIdx > wordIdx; curWordIdx-- {
		prevWordIdx := curWordIdx - 1
		carry := (s.buf[prevWordIdx] >> 63) & 1

		s.buf[curWordIdx] <<= 1
		s.buf[curWordIdx] |= carry
		s.buf[curWordIdx] |= (s.buf[prevWordIdx] << (64 - bitIdx)) >> (64 - bitIdx)
	}

	lowerMask := uint64(1)<<bitIdx - 1
	lowerBits := s.buf[wordIdx] & lowerMask

	upperMask := ^lowerMask
	upperBits := s.buf[wordIdx] & upperMask
	upperBits <<= 1

	s.buf[wordIdx] = upperBits | lowerBits

	if v {
		s.buf[wordIdx] |= 1 << bitIdx
	} else {
		s.buf[wordIdx] &^= 1 << bitIdx
	}
}

func (s *bitslice) getBit(i int) uint8 {
	wordIdx := i / 64
	bitIdx := uint(i % 64)
	return uint8((s.buf[wordIdx] >> bitIdx) & 1)
}

func (s *bitslice) getBool(i int) bool {
	return s.getBit(i) == 1
}

func (s *bitslice) String() string {
	result := "["
	for i := uint64(0); i < s.len; i++ {
		wordIdx := i / 64
		bitIdx := uint(i % 64)
		bit := (s.buf[wordIdx] >> bitIdx) & 1
		if bit == 1 {
			result += "1"
		} else {
			result += "0"
		}
		if i < s.len-1 {
			result += " "
		}
	}
	result += "]"
	return result
}
