package bitvector

import "errors"

var (
	ErrZeroSize         = errors.New("size must be greater than zero")
	ErrInvalidSignature = errors.New("invalid vector signature")
	ErrVersionMismatch  = errors.New("vector version mismatch")
	ErrNotEqualSize     = errors.New("vectors must have equal size")
	ErrWrongType        = errors.New("wrong type provided")
)
