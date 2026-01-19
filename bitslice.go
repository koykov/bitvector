package bitvector

type bitslice struct {
	len uint64
	buf []uint64
}

func (s *bitslice) addBool(v bool) {
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

func (s *bitslice) insertBool(i int, v bool) {
	// todo move to one bit since position i and set value at position i
}
