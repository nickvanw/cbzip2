package cbzip2

import "io"

type Writer struct {
	w   io.Writer
	bz  bzip
	out []byte
	err error
}

// NewWriter returns an io.WriteCloser. Writes to this writer are
// compressed and sent to the underlying writer.
// It is the caller's responsibility to call Close on the WriteCloser.
// Writes may not be flushed until Close.
func NewWriter(w io.Writer) (*Writer, error) {
	wrtr := &Writer{w: w, out: make([]byte, bufferLen)}

	if err := wrtr.bz.compressInit(blockSize, verbosity, workFactor); err != nil {
		return nil, err
	}

	return wrtr, nil
}

// Write writes a compressed p to an underlying io.Writer. The bytes are not
// necessarily flushed until the writer is closed or Flush is called.
func (b *Writer) Write(d []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	b.bz.setInBuf(d, len(d))

	// loop until there's no more input data
	for {
		_, err := b.compress(BZ_RUN)
		if err != nil {
			b.err = err
			return 0, b.err
		}
		// if we've processed all of the input, break
		if b.bz.availIn() == 0 {
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

	_ = b.bz.endCompress()
	b.err = io.EOF
	return nil
}

func (b *Writer) compress(flag int) (int, error) {
	// give the compressor our output buffer
	b.bz.setOutBuf(b.out, len(b.out))

	// add data with our specified call to the buffer
	ret, err := b.bz.compress(flag)
	if err != nil {
		return 0, err
	}

	// we have (total length) - (space available) of data
	have := len(b.out) - b.bz.availOut()
	_, err = b.w.Write(b.out[:have])
	if err != nil {
		_ = b.bz.endCompress()
		return 0, err
	}

	return int(ret), nil
}
