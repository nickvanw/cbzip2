package cbzip2

/*
#cgo LDFLAGS: -lbz2
#cgo CFLAGS: -Werror=implicit

#include <bzlib.h>
*/
import "C"
import (
	"io"
	"unsafe"
)

type bzipReader struct {
	r      io.Reader
	bz     *C.bz_stream
	in     []byte
	skipIn bool
}

// NewBzipWriter returns an io.WriteCloser. Writes to this writer are
// compressed and sent to the underlying writer.
// It is the callers responsibility to call Close on the WriteCloser.
// Writes may not be flushed until Close.
func NewReader(r io.Reader) (io.ReadCloser, error) {
	rdr := &bzipReader{r: r, in: make([]byte, bufferLen)}

	// We dont want to use a custom memory allocator, so we set
	// bzalloc, bzfree and opaque to NULL, to use malloc / free
	rdr.bz = &C.bz_stream{bzalloc: nil, bzfree: nil, opaque: nil}

	if result := C.BZ2_bzDecompressInit(rdr.bz, verbosity, 0); result != BZ_OK {
		return nil, BzipError{Message: "unable to initialize", ReturnCode: int(result)}
	}

	return rdr, nil
}

func (r *bzipReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	var err error
	// read and deflate until the output buffer is full
	r.bz.avail_out = (C.uint)(len(p))
	r.bz.next_out = (*C.char)(unsafe.Pointer(&p[0]))
	for {
		// if the amount of available data to read is 0
		// we reach to the wrapped reader to get more data
		// otherwise, we compress what data is already available
		if !r.skipIn && r.bz.avail_in == 0 {
			var n int
			n, err = r.r.Read(r.in)

			// we are done with reading
			if n == 0 && err == io.EOF {
				C.BZ2_bzDecompressEnd(r.bz)
				return n, err
			}

			// we have data, and EOF
			// disregard the error
			// this will cause a superflous call to Read
			if n > 0 && err == io.EOF {
				err = nil
			}
			if n == 0 && err != nil {
				// if we don't have any data and we errored, close and return
				C.BZ2_bzDecompressEnd(r.bz)
				return 0, err
			}
			// if we do have an error, but we read data, we want to process it
			// and return the error at the bottom

			r.bz.next_in = (*C.char)(unsafe.Pointer(&r.in[0]))
			r.bz.avail_in = (C.uint)(n)
		} else {
			r.skipIn = false // try again
		}
		ret := C.BZ2_bzDecompress(r.bz)
		switch ret {
		case BZ_PARAM_ERROR:
			return 0, BzipError{Message: "param error", ReturnCode: int(ret)}
		case BZ_DATA_ERROR:
			return 0, BzipError{Message: "data integrity error detected", ReturnCode: int(ret)}
		case BZ_DATA_ERROR_MAGIC:
			return 0, BzipError{Message: "compressed stream doesn't begin with the right magic bytes", ReturnCode: int(ret)}
		case BZ_MEM_ERROR:
			return 0, BzipError{Message: "insufficent memory available", ReturnCode: int(ret)}
		}
		// check if we've read anything, if so, return it.
		have := len(p) - int(r.bz.avail_out)
		if have > 0 || err != nil {
			// if the there is no output buffer and we returned OK
			// we want to skip the next read
			r.skipIn = (ret == BZ_OK && r.bz.avail_out == 0)
			return have, err
		}
	}
}

func (r *bzipReader) Close() error {
	C.BZ2_bzDecompressEnd(r.bz)
	return nil
}
