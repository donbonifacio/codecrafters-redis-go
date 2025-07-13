package main

import (
	"bytes"
	"testing"
)

func processCommand(t *testing.T, previous *Connection, cmd string, expectedResponse string) *Connection {
	connection := &Connection{
		reader:   bytes.NewBufferString(cmd),
		writer:   bytes.NewBufferString(""),
		commands: globalCommands,
		KV:       map[string]string{},
	}

	if previous != nil {
		connection.KV = previous.KV
	}

	err := handleConnection(connection)
	if err != nil {
		t.Fatalf("handleConnection returned error: %v", err)
	}
	if connection.response != expectedResponse {
		t.Errorf("expected response = %q, got %q", expectedResponse, connection.response)
	}
	return connection
}

func TestSetGet(t *testing.T) {
	connection := processCommand(t, nil,
		"*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
		"+OK\r\n",
	)
	processCommand(t, connection,
		"*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n",
		"$5\r\nvalue\r\n",
	)
}

func TestReadRedisCmd(t *testing.T) {
	// Example Redis command: "*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n"
	input := "*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n"
	r := bytes.NewBufferString(input)

	cmd, err := readRedisCmd(r)
	if err != nil {
		t.Fatalf("readRedisCmd returned error: %v", err)
	}

	expected := []string{"PING", "PONG"}
	if len(cmd) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(cmd))
	}
	for i := range expected {
		if cmd[i] != expected[i] {
			t.Errorf("expected cmd[%d] = %q, got %q", i, expected[i], cmd[i])
		}
	}
}
