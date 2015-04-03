package cbzip2

/*
#cgo LDFLAGS: -lbz2
#cgo CFLAGS: -Werror=implicit

#include "bzip2/bzlib.h"
*/
import "C"
import (
	"io"
	"unsafe"
)

type Reader struct {
	r      io.Reader
	bz     *C.bz_stream
	in     []byte
	skipIn bool
	err    error
}

// NewReader returns an io.ReadCloser. Reads from this are read from the
// underlying io.Reader and decompressed via bzip2
func NewReader(r io.Reader) (*Reader, error) {
	rdr := &Reader{r: r, in: make([]byte, bufferLen)}

	// We dont want to use a custom memory allocator, so we set
	// bzalloc, bzfree and opaque to NULL, to use malloc / free
	rdr.bz = &C.bz_stream{bzalloc: nil, bzfree: nil, opaque: nil}

	if result := C.BZ2_bzDecompressInit(rdr.bz, verbosity, 0); result != BZ_OK {
		return nil, ErrInit
	}

	return rdr, nil
}

// Read pulls data up from the underlying io.Reader and decompresses the data
func (r *Reader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	// read and deflate until the output buffer is full
	r.bz.avail_out = (C.uint)(len(p))
	r.bz.next_out = (*C.char)(unsafe.Pointer(&p[0]))
	for {
		// if the amount of available data to read is 0
		// we reach to the wrapped reader to get more data
		// otherwise, we compress what data is already available
		if !r.skipIn && r.bz.avail_in == 0 {
			var n int
			n, r.err = r.r.Read(r.in)

			// we are done with reading
			if n == 0 && r.err == io.EOF {
				C.BZ2_bzDecompressEnd(r.bz)
				return n, r.err
			}

			// we have data, and EOF
			// disregard the error
			// this will cause a superflous call to Read
			if n > 0 && r.err == io.EOF {
				r.err = nil
			}
			if n == 0 && r.err != nil {
				// if we don't have any data and we errored, close and return
				C.BZ2_bzDecompressEnd(r.bz)
				return 0, r.err
			}
			// if we do have an error, but we read data, we want to process it
			// and return the error at the bottom

			r.bz.next_in = (*C.char)(unsafe.Pointer(&r.in[0]))
			r.bz.avail_in = (C.uint)(n)
		} else {
			r.skipIn = false // try again
		}
		ret := C.BZ2_bzDecompress(r.bz)
		var err error
		switch ret {
		case BZ_PARAM_ERROR:
			err = ErrBadParam
		case BZ_DATA_ERROR:
			err = ErrBadData
		case BZ_DATA_ERROR_MAGIC:
			err = ErrBadMagic
		case BZ_MEM_ERROR:
			err = ErrMem
		}
		if err != nil {
			r.err = err
		}
		// check if we've read anything, if so, return it.
		have := len(p) - int(r.bz.avail_out)
		if have > 0 || r.err != nil {
			// if the there is no output buffer and we returned OK
			// we want to skip the next read
			r.skipIn = (ret == BZ_OK && r.bz.avail_out == 0)
			return have, r.err
		}
	}
}

// Close closes the reader, but not the underlying io.Reader
func (r *Reader) Close() error {
	if r.err != nil {
		return r.err
	}
	C.BZ2_bzDecompressEnd(r.bz)
	r.err = io.EOF
	return nil
}
