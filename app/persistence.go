package main

import (
	"fmt"
	"io"
	"os"
)

type LoadDBResult struct {
	header string
}

func loadDB(config map[string]string) (map[string]RedisValue, *LoadDBResult, error) {
	db := map[string]RedisValue{}

	filename := fmt.Sprintf("%s/%s", config["dir"], config["dbfilename"])
	if filename == "/" {
		return db, nil, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	header, err := readHeader(file)
	if err != nil {
		return nil, nil, err
	}

	return db, &LoadDBResult{header}, nil
}

func readHeader(file io.Reader) (string, error) {
	buf := make([]byte, 9)
	n, err := file.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func readMetadataItem(r io.Reader) (name string, value string, err error) {
	buf := make([]byte, 1)
	n, err := r.Read(buf)
	if err != nil {
		return "", "", err
	}
	if n != 1 {
		return "", "", fmt.Errorf("expected 1 byte for metadata marker, got %d", n)
	}
	if buf[0] != 0xFA {
		return "", "", fmt.Errorf("expected metadata marker 0xFA, got 0x%X", buf[0])
	}

	nameResult, err := readEncoded(r)
	if err != nil {
		return "", "", fmt.Errorf("error reading metadata name: %w", err)
	}

	valueResult, err := readEncoded(r)
	if err != nil {
		return "", "", fmt.Errorf("error reading metadata value: %w", err)
	}

	return nameResult.stringValue, valueResult.stringValue, nil
}

type ReadBlock struct {
	name             string
	mask             uint8
	extraBytesToRead func(b byte) int
	decodeString     func(bytes []byte) string
	decodeInt        func(bytes []byte) int
}

type ReadBlockResult struct {
	stringValue      string
	intValue         int
	returnType       string
	bytes            []byte
	extraBytesToRead int
}

func (r *ReadBlockResult) toString() string {
	return fmt.Sprintf(
		"ReadBlockResult{stringValue: %q, intValue: %d, returnType: %q, bytes: %v, extraBytesToRead: %d}",
		r.stringValue, r.intValue, r.returnType, r.bytes, r.extraBytesToRead,
	)
}

func readEncoded(r io.Reader) (*ReadBlockResult, error) {
	blocks := []ReadBlock{
		ReadBlock{
			name: "int8",
			mask: 0xC0,
			extraBytesToRead: func(_ byte) int {
				return 1
			},
			decodeString: func(bytes []byte) string {
				return fmt.Sprintf("%d", int(bytes[1]))
			},
		},
		ReadBlock{
			name: "int16",
			mask: 0xC1,
			extraBytesToRead: func(_ byte) int {
				return 2
			},
			decodeString: func(bytes []byte) string {
				return fmt.Sprintf("%d", int(int16(bytes[1])|(int16(bytes[2])<<8)))
			},
		},
		ReadBlock{
			name: "int24",
			mask: 0xC2,
			extraBytesToRead: func(_ byte) int {
				return 4
			},
			decodeString: func(bytes []byte) string {
				return fmt.Sprintf("%d", int(uint32(bytes[1])|(uint32(bytes[2])<<8)|(uint32(bytes[3])<<16)|(uint32(bytes[4])<<24)))
			},
		},
		// catch all
		ReadBlock{
			name: "string",
			extraBytesToRead: func(b byte) int {
				return int(b)
			},
			decodeString: func(bytes []byte) string {
				return string(bytes[1:])
			},
		},
	}
	buf := make([]byte, 1)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, fmt.Errorf("expected 1 bytes for length, got %d", n)
	}
	var block *ReadBlock
	for _, b := range blocks {
		if buf[0]&b.mask == buf[0] {
			block = &b
			break
		}
	}
	if block == nil {
		// last one, catch all
		block = &blocks[len(blocks)-1]
	}
	fmt.Printf("block: %v\n", block)
	extraBytesToRead := block.extraBytesToRead(buf[0])
	strBuf := make([]byte, extraBytesToRead)
	n, err = r.Read(strBuf)
	if err != nil {
		return nil, err
	}
	if n != extraBytesToRead {
		return nil, fmt.Errorf("expected %d bytes for string, got %d", extraBytesToRead, n)
	}
	strBuf = append([]byte{buf[0]}, strBuf...)
	if block.decodeString != nil {
		return &ReadBlockResult{
			bytes:            strBuf,
			extraBytesToRead: extraBytesToRead,
			stringValue:      block.decodeString(strBuf),
			returnType:       "string",
		}, nil
	}
	if block.decodeInt != nil {
		return &ReadBlockResult{
			extraBytesToRead: extraBytesToRead,
			intValue:         block.decodeInt(strBuf),
			bytes:            strBuf,
			returnType:       "int",
		}, nil
	}
	return nil, fmt.Errorf("unsupported block type")
}
