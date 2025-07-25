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
