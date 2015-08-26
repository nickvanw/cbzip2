package cbzip2

import "io"

type Reader struct {
	r      io.Reader
	bz     bzip
	in     []byte
	skipIn bool
	err    error
}

// NewReader returns an io.ReadCloser. Reads from this are read from the
// underlying io.Reader and decompressed via bzip2
func NewReader(r io.Reader) (*Reader, error) {
	rdr := &Reader{r: r, in: make([]byte, bufferLen)}

	if err := rdr.bz.decompressInit(verbosity, 0); err != nil {
		return nil, err
	}
	return rdr, nil
}

// Read pulls data up from the underlying io.Reader and decompresses the data
func (r *Reader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	// read and deflate until the output buffer is full
	r.bz.setOutBuf(p, len(p))
	for {
		// if the amount of available data to read is 0
		// we reach to the wrapped reader to get more data
		// otherwise, we compress what data is already available
		if !r.skipIn && r.bz.availIn() == 0 {
			var n int
			n, r.err = r.r.Read(r.in)

			// we are done with reading
			if n == 0 && r.err == io.EOF {
				_ = r.bz.endDecompress()
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
				_ = r.bz.endDecompress()
				return 0, r.err
			}
			// if we do have an error, but we read data, we want to process it
			// and return the error at the bottom
			r.bz.setInBuf(r.in, n)
		} else {
			r.skipIn = false // try again
		}
		ret := r.bz.decompress()
		if ret < 0 {
			r.err = retCodeToErr(int(ret))
		}
		// check if we've read anything, if so, return it.
		have := len(p) - int(r.bz.availOut())
		if have > 0 || r.err != nil {
			// if the there is no output buffer and we returned OK
			// we want to skip the next read
			r.skipIn = (ret == BZ_OK && r.bz.availOut() == 0)
			return have, r.err
		}
	}
}

// Close closes the reader, but not the underlying io.Reader
func (r *Reader) Close() error {
	if r.err != nil {
		return r.err
	}
	_ = r.bz.endDecompress()
	r.err = io.EOF
	return nil
}
