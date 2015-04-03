package cbzip2

import "errors"

var (
	ErrBadParam       = errors.New("bad param (should be impossible)")
	ErrBadData        = errors.New("integrity problem detected in input data")
	ErrBadMagic       = errors.New("compressed stream does not being with magic bytes")
	ErrMem            = errors.New("insufficient memory available ಠ_ಠ")
	ErrInit           = errors.New("unable to initialize bzlib.h")
	ErrBadCompression = errors.New("unable to compress data")
)
