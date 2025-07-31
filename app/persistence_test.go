package main

import (
	"bytes"
	"testing"
)

func TestCheckHeyKeyExists(t *testing.T) {
	config := map[string]string{}
	config["dir"] = ".."
	config["dbfilename"] = "dump.rdb"

	mainDB, loadDBResult, err := loadDB(config)
	if err != nil {
		t.Fatal(err)
	}

	if loadDBResult == nil {
		t.Fatal("loadDBResult is nil")
	}

	if string(loadDBResult.header) != string("REDIS0011") {
		t.Errorf("expected header %v, got %v", "REDIS0011", loadDBResult.header)
	}

	connection := createTestConnection(config)
	connection.KV = mainDB

	processCommand(t, connection,
		resp("GET", "hey"),
		resp("hey"),
	)
}
func TestReadHeader(t *testing.T) {
	header := "REDIS0006"
	expectedHeader := []byte(header)
	reader := bytes.NewReader(expectedHeader)

	header, err := readHeader(reader)
	if err != nil {
		t.Fatalf("readHeader returned error: %v", err)
	}

	if string(header) != string(expectedHeader) {
		t.Errorf("expected header %q, got %q", expectedHeader, header)
	}
}

func TestReadMetadataEntry(t *testing.T) {
	// Example metadata entry: "redis-ver" = "6.0.16"
	metadataEntry := []byte{
		0xFA,                                                       // Start of metadata subsection
		0x09, 0x72, 0x65, 0x64, 0x69, 0x73, 0x2D, 0x76, 0x65, 0x72, // "redis-ver"
		0x06, 0x36, 0x2E, 0x30, 0x2E, 0x31, 0x36, // "6.0.16"
	}
	reader := bytes.NewReader(metadataEntry)
	name, value, err := readMetadataItem(reader)
	if err != nil {
		t.Fatalf("readHeader returned error: %v", err)
	}

	if name != "redis-ver" {
		t.Errorf("expected name %v, got %v", name, "redis-ver")
	}
	if value != "6.0.16" {
		t.Errorf("expected name %v, got %v", value, "6.0.16")
	}
}

func checkEncodedString(t *testing.T, encoded []byte, expected string) {
	reader := bytes.NewReader(encoded)
	result, err := readEncodedString(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != expected {
		t.Errorf("expected header %v, got %v", expected, result)
	}
}

func TestReadEncodedString(t *testing.T) {
	encoded := []byte{0x0D, 0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x2C, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, 0x21}
	expected := "Hello, World!"
	checkEncodedString(t, encoded, expected)
}

func TestReadEncodedString8bitint(t *testing.T) {
	encoded := []byte{0xC0, 0x7B}
	expected := "123"
	checkEncodedString(t, encoded, expected)
}

func TestReadEncodedString16bitint(t *testing.T) {
	encoded := []byte{0xC1, 0x39, 0x30}
	expected := "12345"
	checkEncodedString(t, encoded, expected)
}

func TestReadEncodedString32bitint(t *testing.T) {
	encoded := []byte{0xC2, 0x87, 0xD6, 0x12, 0x00}
	expected := "1234567"
	checkEncodedString(t, encoded, expected)
}
