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

type bzipWriter struct {
	w   io.Writer
	bz  *C.bz_stream
	out []byte
}

// NewWriter returns an io.WriteCloser. Writes to this writer are
// compressed and sent to the underlying writer.
// It is the callers responsibility to call Close on the WriteCloser.
// Writes may not be flushed until Close.
func NewWriter(w io.Writer) (io.WriteCloser, error) {
	wrtr := &bzipWriter{w: w, out: make([]byte, bufferLen)}

	// We dont want to use a custom memory allocator, so we set
	// bzalloc, bzfree and opaque to NULL, to use malloc / free
	wrtr.bz = &C.bz_stream{bzalloc: nil, bzfree: nil, opaque: nil}

	if result := C.BZ2_bzCompressInit(wrtr.bz, blockSize, verbosity, workFactor); result != BZ_OK {
		return nil, BzipError{Message: "unable to initialize", ReturnCode: int(result)}
	}

	return wrtr, nil
}

// Write writes a compressed p to an underlying io.Writer. The bytes are not
// necessarily flushed until the writer is closed or Flush is called.
func (b *bzipWriter) Write(d []byte) (int, error) {
	return b.write(d, BZ_RUN)
}

// Flush writes any pending data to the underlying writer.
func (b *bzipWriter) Flush() error {
	_, err := b.write(nil, BZ_FLUSH)
	return err
}

// Close closes the writer, flushing any unwritten data to the underlying io.Writer
// Close does not close the underlying io.Writer.
func (b *bzipWriter) Close() error {
	if _, err := b.write(nil, BZ_FINISH); err != nil {
		return err
	}
	C.BZ2_bzCompressEnd(b.bz)
	return io.EOF
}

func (b *bzipWriter) write(d []byte, flush int) (int, error) {
	if len(d) == 0 {
		b.bz.avail_in = 0
		b.bz.next_in = (*C.char)(unsafe.Pointer(nil))
	} else {
		b.bz.avail_in = (C.uint)(len(d))
		b.bz.next_in = (*C.char)(unsafe.Pointer(&d[0]))
	}

	// loop until we don't have a full output buffer
	// this will also write to the underlying writer
	for {
		// give the compressor our output buffer
		b.bz.next_out = (*C.char)(unsafe.Pointer(&b.out[0]))
		b.bz.avail_out = (C.uint)(len(b.out))
		// add data with our specified call to the buffer
		if ret := C.BZ2_bzCompress(b.bz, (C.int)(flush)); ret < 0 {
			return 0, BzipError{Message: "unable to compress", ReturnCode: int(ret)}
		}
		// we have (total length) - (space available) of data
		have := len(b.out) - int(b.bz.avail_out)
		_, err := b.w.Write(b.out[:have])
		if err != nil {
			C.BZ2_bzCompressEnd(b.bz)
			return 0, err
		}
		// we have available output buffer, drop out
		// and get more data
		if b.bz.avail_out != 0 {
			break
		}
	}
	return len(d), nil
}
