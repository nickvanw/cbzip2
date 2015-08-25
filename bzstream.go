package cbzip2

/*
#cgo CFLAGS: -Werror=implicit

#include "bzlib.h"

int bz_compress_init(char *strm, int blockSize, int verbosity, int workFactor) {
	((bz_stream*)strm)->bzalloc = NULL;
	((bz_stream*)strm)->bzfree = NULL;
	((bz_stream*)strm)->opaque = NULL;
	return BZ2_bzCompressInit((bz_stream*)strm,
	                           blockSize, verbosity, workFactor);
}

int bz_decompress_init(char *strm, int verbosity, int small) {
	((bz_stream*)strm)->bzalloc = NULL;
	((bz_stream*)strm)->bzfree = NULL;
	((bz_stream*)strm)->opaque = NULL;
	return BZ2_bzDecompressInit((bz_stream*)strm,
	                           verbosity, small);
}

unsigned int stream_avail_in(char *strm) {
	return ((bz_stream*)strm)->avail_in;
}

unsigned int stream_avail_out(char *strm) {
	return ((bz_stream*)strm)->avail_out;
}

void stream_set_in_buf(char *strm, char *buf, unsigned int len) {
	((bz_stream*)strm)->next_in = buf;
	((bz_stream*)strm)->avail_in = len;
}

void stream_set_out_buf(char *strm, char *buf, unsigned int len) {
	((bz_stream*)strm)->next_out = buf;
	((bz_stream*)strm)->avail_out = len;
}

int stream_compress(char *strm, int flag) {
	return BZ2_bzCompress((bz_stream*)strm, flag);
}

int stream_decompress(char *strm) {
	return BZ2_bzDecompress((bz_stream*)strm);
}

int stream_compress_end(char *strm) {
	return BZ2_bzCompressEnd((bz_stream*)strm);
}

int stream_decompress_end(char *strm) {
	return BZ2_bzDecompressEnd((bz_stream*)strm);
}
*/
import "C"
import "unsafe"

type bzip [unsafe.Sizeof(C.bz_stream{})]C.char

func (b *bzip) compressInit(blockSize, verbosity, workFactor int) error {
	if result := C.bz_compress_init(&b[0], C.int(blockSize), C.int(verbosity), C.int(workFactor)); result != BZ_OK {
		return retCodeToErr(int(result))
	}
	return nil
}

func (b *bzip) decompressInit(verbosity, small int) error {
	if result := C.bz_decompress_init(&b[0], C.int(verbosity), C.int(small)); result != BZ_OK {
		return retCodeToErr(int(result))
	}
	return nil
}

func (b *bzip) availIn() int {
	return int(C.stream_avail_in(&b[0]))
}

func (b *bzip) setInBuf(buf []byte, size int) {
	if buf == nil || len(buf) == 0 {
		C.stream_set_in_buf(&b[0], nil, C.uint(size))
	} else {
		C.stream_set_in_buf(&b[0], (*C.char)(unsafe.Pointer(&buf[0])), C.uint(size))
	}
}

func (b *bzip) availOut() int {
	return int(C.stream_avail_out(&b[0]))
}

func (b *bzip) setOutBuf(buf []byte, size int) {
	if buf == nil || len(buf) == 0 {
		C.stream_set_out_buf(&b[0], nil, C.uint(size))
	} else {
		C.stream_set_out_buf(&b[0], (*C.char)(unsafe.Pointer(&buf[0])), C.uint(size))
	}
}

func (b *bzip) compress(flag int) (int, error) {
	ret := C.stream_compress(&b[0], C.int(flag))
	if ret < 0 {
		return 0, retCodeToErr(int(ret))
	}
	return int(ret), nil
}

func (b *bzip) decompress() int {
	ret := C.stream_decompress(&b[0])
	return int(ret)
}

func (b *bzip) endCompress() int {
	return int(C.stream_compress_end(&b[0]))
}

func (b *bzip) endDecompress() int {
	return int(C.stream_decompress_end(&b[0]))
}
