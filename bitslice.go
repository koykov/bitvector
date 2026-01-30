package bitvector

import (
	"bytes"
	"encoding/binary"
	"io"
	"unsafe"
)

type bitslice struct {
	ln  uint64
	buf []uint64
}

func (s *bitslice) add(v bool) {
	if s.ln%64 == 0 {
		s.buf = append(s.buf, 0)
	}
	s.ln++
	if !v {
		return
	}
	i, off := s.ln/64, s.ln%64
	s.buf[i] |= 1 << off
}

func (s *bitslice) insert(i int, v bool) {
	if i < 0 || uint64(i) > s.ln {
		return
	}

	s.ln++

	neededBits := (s.ln + 63) / 64
	if neededBits > uint64(len(s.buf)) {
		newBuf := make([]uint64, neededBits)
		copy(newBuf, s.buf)
		s.buf = newBuf
	}

	wordIdx := uint64(i) / 64
	bitIdx := uint64(i) % 64

	if uint64(i) == s.ln-1 {
		if v {
			s.buf[wordIdx] |= 1 << bitIdx
		} else {
			s.buf[wordIdx] &^= 1 << bitIdx
		}
		return
	}

	lastWordIdx := (s.ln - 1) / 64

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

func (s *bitslice) delete(i int) bool {
	if i < 0 || uint64(i) >= s.ln {
		return false
	}

	if s.ln == 0 {
		return false
	}

	wordIdx := uint64(i) / 64
	bitIdx := uint64(i) % 64

	lastWordIdx := (s.ln - 1) / 64

	maskBefore := uint64(1) << bitIdx
	maskBefore = maskBefore - 1

	maskAfter := ^((uint64(1) << (bitIdx + 1)) - 1)

	lowerBits := s.buf[wordIdx] & maskBefore

	upperBits := s.buf[wordIdx] & maskAfter
	upperBits >>= 1

	s.buf[wordIdx] = lowerBits | upperBits

	for curWordIdx := wordIdx; curWordIdx < lastWordIdx; curWordIdx++ {

		nextWordIdx := curWordIdx + 1
		lsbFromNext := s.buf[nextWordIdx] & 1

		s.buf[curWordIdx] |= lsbFromNext << 63

		s.buf[nextWordIdx] >>= 1
	}

	s.ln--
	if s.ln > 0 {
		lastWordIdxAfter := (s.ln - 1) / 64
		bitsInLastWord := s.ln % 64
		if bitsInLastWord == 0 {
			bitsInLastWord = 64
		}

		mask := (uint64(1) << bitsInLastWord) - 1
		s.buf[lastWordIdxAfter] &= mask

		for i := lastWordIdxAfter + 1; i < uint64(len(s.buf)); i++ {
			s.buf[i] = 0
		}
	}
	return true
}

func (s *bitslice) getBit(i int) uint8 {
	wordIdx := i / 64
	bitIdx := uint(i % 64)
	return uint8((s.buf[wordIdx] >> bitIdx) & 1)
}

func (s *bitslice) get(i int) bool {
	return s.getBit(i) == 1
}

func (s *bitslice) len() uint64 {
	return s.ln
}

func (s *bitslice) writeTo(w io.Writer) (n int64, err error) {
	var buf [16]byte
	binary.LittleEndian.PutUint64(buf[0:8], s.ln)
	binary.LittleEndian.PutUint64(buf[8:16], uint64(len(s.buf)))
	var n1 int
	if n1, err = w.Write(buf[:]); err != nil {
		return
	}
	n += int64(n1)

	if len(s.buf) == 0 {
		return
	}

	type h struct {
		p    uintptr
		l, c int
	}
	h1 := *(*h)(unsafe.Pointer(&s.buf))
	h1.l *= 8
	h1.c *= 8
	buf1 := *(*[]byte)(unsafe.Pointer(&h1))
	if n1, err = w.Write(buf1); err != nil {
		return
	}
	n += int64(n1)

	return
}

func (s *bitslice) readFrom(r io.Reader) (n int64, err error) {
	// todo implement me
	return
}

func (s *bitslice) reset() {
	s.buf = s.buf[:0]
	s.ln = 0
}

func (s *bitslice) clone() bitslice {
	return bitslice{
		ln:  s.ln,
		buf: append([]uint64{}, s.buf...),
	}
}

func (s *bitslice) String() string {
	var buf bytes.Buffer
	buf.Grow(int(s.ln + 2))
	_ = buf.WriteByte('[')
	for i := uint64(0); i < s.ln; i++ {
		wordIdx := i / 64
		bitIdx := uint(i % 64)
		bit := (s.buf[wordIdx] >> bitIdx) & 1
		if bit == 1 {
			buf.WriteByte('1')
		} else {
			buf.WriteByte('0')
		}
	}
	buf.WriteByte(']')
	return buf.String()
}
