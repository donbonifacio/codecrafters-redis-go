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

	name, err = readEncodedString(r)
	if err != nil {
		return "", "", fmt.Errorf("error reading metadata name: %w", err)
	}

	value, err = readEncodedString(r)
	if err != nil {
		return "", "", fmt.Errorf("error reading metadata value: %w", err)
	}

	return name, value, nil
}

func readEncodedString(r io.Reader) (string, error) {
	buf := make([]byte, 1)
	n, err := r.Read(buf)
	if err != nil {
		return "", err
	}
	if n != 1 {
		return "", fmt.Errorf("expected 1 bytes for length, got %d", n)
	}
	length := parseLength(buf[0])
	strBuf := make([]byte, length)
	n, err = r.Read(strBuf)
	fmt.Printf("------%v %v\n", strBuf, string(strBuf))
	if err != nil {
		return "", err
	}
	if n != length {
		return "", fmt.Errorf("expected %d bytes for string, got %d", length, n)
	}
	return parseContent(buf[0], strBuf), nil
}

func parseLength(b byte) int {
	if b == 0xC0 {
		// 8 bit int
		return 1
	}
	if b == 0xC1 {
		// 16 bit int
		return 2
	}
	if b == 0xC2 {
		// 32 bit int
		return 4
	}
	return int(b)

}

func parseContent(b byte, strBuf []byte) string {
	if b == 0xC0 {
		// 8 bit int
		return fmt.Sprintf("%d", int(strBuf[0]))
	}
	if b == 0xC1 {
		// 16 bit int
		return fmt.Sprintf("%d", int(int16(strBuf[0])|(int16(strBuf[1])<<8)))
	}
	if b == 0xC2 {
		// 32 bit int
		return fmt.Sprintf("%d", int(uint32(strBuf[0])|(uint32(strBuf[1])<<8)|(uint32(strBuf[2])<<16)|(uint32(strBuf[3])<<24)))
	}
	return string(strBuf)
}
