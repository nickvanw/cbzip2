package cbzip2

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
