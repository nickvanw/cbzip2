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

const (
	// blockSize specifies the block size to be used for compression.
	// It should be a value between 1 and 9 inclusive, and the actual block size
	// used is 100000 x this figure. 9 gives the best compression but takes most memory.
	blockSize = 9 // default
	// verbosity should be set to a number between 0 and 4 inclusive.
	// 0 is silent, and greater numbers give increasingly verbose monitoring/debugging output.
	verbosity = 0
	// workFactor is the amount of effort the standard algorithm will expend before
	// resorting to the fallback.
	workFactor = 30 // default

	// bufferLen is our default buffer size, set to 32KB which is common for other io functions
	bufferLen = 32 * 1024
)

// copied from bzlib.h
const (
	BZ_RUN    = 0
	BZ_FLUSH  = 1
	BZ_FINISH = 2

	// return codes
	BZ_OK               = 0
	BZ_RUN_OK           = 1
	BZ_FLUSH_OK         = 2
	BZ_FINISH_OK        = 3
	BZ_STREAM_END       = 4
	BZ_SEQUENCE_ERROR   = (-1)
	BZ_PARAM_ERROR      = (-2)
	BZ_MEM_ERROR        = (-3)
	BZ_DATA_ERROR       = (-4)
	BZ_DATA_ERROR_MAGIC = (-5)
	BZ_IO_ERROR         = (-6)
	BZ_UNEXPECTED_EOF   = (-7)
	BZ_OUTBUFF_FULL     = (-8)
	BZ_CONFIG_ERROR     = (-9)
)

type bzipWriter struct {
	w   io.Writer
	bz  *C.bz_stream
	out []byte
}

func NewBzipWriter(w io.Writer) (io.WriteCloser, error) {
	wrtr := &bzipWriter{w: w, out: make([]byte, bufferLen)}

	// We dont want to use a custom memory allocator, so we set
	// bzalloc, bzfree and opaque to NULL, to use malloc / free
	wrtr.bz = &C.bz_stream{bzalloc: nil, bzfree: nil, opaque: nil}

	if result := C.BZ2_bzCompressInit(wrtr.bz, blockSize, verbosity, workFactor); result != BZ_OK {
		return nil, BzipError{Message: "unable to initialize", ReturnCode: int(result)}
	}

	return wrtr, nil
}

func (b *bzipWriter) Write(d []byte) (int, error) {
	return b.write(d, BZ_RUN)
}

func (b *bzipWriter) Flush() error {
	_, err := b.write(nil, BZ_FLUSH)
	return err
}

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

	// loop until we dont have a full output buffer
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
