
# cbzip2
    import "github.com/nickvanw/cbzip2"

package cbzip2 provides access to the "low level" bzip2 interface
via bzlib.h, documented here: a href="http://www.bzip.org/1.0.3/html/low-level.html">http://www.bzip.org/1.0.3/html/low-level.html</a>




## Constants
``` go
const (
    // valid actions
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
```
copied from bzlib.h



## func NewBzipWriter
``` go
func NewBzipWriter(w io.Writer) (io.WriteCloser, error)
```
NewBzipWriter returns an io.WriteCloser. Writes to this writer are
compressed and sent to the underlying writer.
It is the callers responsibility to call Close on the WriteCloser.
Writes may not be flushed until Close.



## type BzipError
``` go
type BzipError struct {
    ReturnCode int
    Message    string
}
```
BzipError represents an error returned during operation
of bzlib. It contains a message about the attempted action
as well as the bzlib return code.











### func (BzipError) Error
``` go
func (e BzipError) Error() string
```








- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
