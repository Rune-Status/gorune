package gorune

import (
	"compress/flate"
	"fmt"

	"encoding/binary"

	"bytes"

	"io/ioutil"

	"compress/bzip2"
)

// MaximumRawSize holds the maximum accepted size when decompressing
// a compressed entry. If the decompressed size exceeds this number,
// it is not further processed and assumed to be a faulty chunk of data.
var MaximumRawSize = 15000000

type CompressionType int

const (
	NoCompression   CompressionType = 0
	Bzip2Compresion                 = 1
	GzipCompression                 = 2
)

func Decompress(compressed []byte) ([]byte, CompressionType, error) {
	if len(compressed) < 5 {
		return nil, NoCompression,
			fmt.Errorf("cannot decompress data because the byte array (%db) has less than 5 bytes", len(compressed))
	}

	compression := CompressionType(compressed[0])
	compressedSize := binary.BigEndian.Uint32(compressed[1:])

	// If the compression type is 0, we return the data.
	if compression == NoCompression {
		return compressed[5 : compressedSize+5], compression, nil
	}

	decompressedSize := binary.BigEndian.Uint32(compressed[5:])

	if decompressedSize > uint32(MaximumRawSize) {
		return nil, compression,
			fmt.Errorf("refusing to decompress data because %d is more than %d", decompressedSize, MaximumRawSize)
	}

	if compression == Bzip2Compresion {
		decompress, err := decompressBzip2(compressed[9:])
		return decompress, compression, err
	}

	if compression == GzipCompression {
		decompressed, err := decompressGzip(compressed[9:])
		return decompressed, compression, err
	}

	return nil, compression, nil
}

func decompressGzip(data []byte) ([]byte, error) {
	// Make sure this archive has a valid Gzip header
	if data[0] != 0x1F || data[1] != 0x8B {
		return nil, fmt.Errorf("invalid gzip header: %x %x (expected 0x1F 0x8B)", data[0], data[1])
	}

	buffer := bytes.NewReader(data[10 : len(data)-4])
	reader := flate.NewReader(buffer)
	return ioutil.ReadAll(reader)
}

func decompressBzip2(data []byte) ([]byte, error) {
	header := []byte{'B', 'Z', 'h', '1'} // Required Bzip2 header
	buffer := bytes.NewReader(append(header, data[:len(data)-2]...))
	bzipper := bzip2.NewReader(buffer)
	return ioutil.ReadAll(bzipper)
}
