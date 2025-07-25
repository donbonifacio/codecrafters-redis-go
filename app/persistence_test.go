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
