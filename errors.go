package cbzip2

import "errors"

var (
	ErrBadParam = errors.New("bad parameters given to bzip")
	ErrBadData  = errors.New("integrity problem detected in input data")
	ErrBadMagic = errors.New("compressed stream does not being with magic bytes")
	ErrMem      = errors.New("insufficient memory available ಠ_ಠ")
	ErrSequence = errors.New("bzip2 sequence error")
	ErrIo       = errors.New("i/o error")
	ErrEOF      = errors.New("unexpected EOF")
	ErrOutFull  = errors.New("output buffer full")
	ErrConfig   = errors.New("config error")
	ErrUnknown  = errors.New("unknown error")
)

func retCodeToErr(ret int) error {
	switch ret {
	case BZ_SEQUENCE_ERROR:
		return ErrSequence
	case BZ_PARAM_ERROR:
		return ErrBadParam
	case BZ_MEM_ERROR:
		return ErrMem
	case BZ_DATA_ERROR:
		return ErrBadData
	case BZ_DATA_ERROR_MAGIC:
		return ErrBadMagic
	case BZ_IO_ERROR:
		return ErrIo
	case BZ_UNEXPECTED_EOF:
		return ErrEOF
	case BZ_OUTBUFF_FULL:
		return ErrOutFull
	case BZ_CONFIG_ERROR:
		return ErrConfig
	default:
		return ErrUnknown
	}
}
