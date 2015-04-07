package cbzip2

/*
#cgo CFLAGS: -Werror=implicit

#include "bzlib.h"
*/
import "C"
import (
	"io"
	"unsafe"
)

type Writer struct {
	w   io.Writer
	bz  *C.bz_stream
	out []byte
	err error
}

// NewWriter returns an io.WriteCloser. Writes to this writer are
// compressed and sent to the underlying writer.
// It is the caller's responsibility to call Close on the WriteCloser.
// Writes may not be flushed until Close.
func NewWriter(w io.Writer) (*Writer, error) {
	wrtr := &Writer{w: w, out: make([]byte, bufferLen)}

	// We dont want to use a custom memory allocator, so we set
	// bzalloc, bzfree and opaque to NULL, to use malloc / free
	wrtr.bz = &C.bz_stream{bzalloc: nil, bzfree: nil, opaque: nil}

	if result := C.BZ2_bzCompressInit(wrtr.bz, blockSize, verbosity, workFactor); result != BZ_OK {
		return nil, retCodeToErr(int(result))
	}

	return wrtr, nil
}

// Write writes a compressed p to an underlying io.Writer. The bytes are not
// necessarily flushed until the writer is closed or Flush is called.
func (b *Writer) Write(d []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	if len(d) == 0 {
		b.bz.avail_in = 0
		b.bz.next_in = (*C.char)(unsafe.Pointer(nil))
	} else {
		b.bz.avail_in = (C.uint)(len(d))
		b.bz.next_in = (*C.char)(unsafe.Pointer(&d[0]))
	}

	// loop until there's no more input data
	for {
		_, err := b.compress(BZ_RUN)
		if err != nil {
			b.err = err
			return 0, b.err
		}
		// if we've processed all of the input, break
		if b.bz.avail_in == 0 {
			break
		}
	}
	return len(d), nil
}

// Flush writes any pending data to the underlying writer.
func (b *Writer) Flush() error {
	if b.err != nil {
		return b.err
	}
	for {
		ret, err := b.compress(BZ_FLUSH)
		if err != nil {
			b.err = err
			return b.err
		}
		// if we're done flushing, return
		if ret == BZ_RUN_OK {
			break
		}
	}
	return nil
}

// Close closes the writer, flushing any unwritten data to the underlying io.Writer
// Close does not close the underlying io.Writer.
func (b *Writer) Close() error {
	if b.err != nil {
		return b.err
	}
	for {
		ret, err := b.compress(BZ_FINISH)
		if err != nil {
			b.err = err
			return b.err
		}
		// When we get to the actual end of the stream, break
		if ret == BZ_STREAM_END {
			break
		}
	}

	C.BZ2_bzCompressEnd(b.bz)
	b.err = io.EOF
	return nil
}

func (b *Writer) compress(flush int) (int, error) {
	// give the compressor our output buffer
	b.bz.next_out = (*C.char)(unsafe.Pointer(&b.out[0]))
	b.bz.avail_out = (C.uint)(len(b.out))

	// add data with our specified call to the buffer
	ret := C.BZ2_bzCompress(b.bz, (C.int)(flush))
	if ret < 0 {
		return 0, retCodeToErr(int(ret))
	}

	// we have (total length) - (space available) of data
	have := len(b.out) - int(b.bz.avail_out)
	_, err := b.w.Write(b.out[:have])
	if err != nil {
		C.BZ2_bzCompressEnd(b.bz)
		return 0, err
	}

	return int(ret), nil
}
