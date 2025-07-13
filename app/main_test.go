package main

import (
	"bytes"
	"testing"
)

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
