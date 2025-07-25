package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func createTestConnection(rawConfig map[string]string) *Connection {
	if rawConfig == nil {
		rawConfig = map[string]string{}
	}
	return &Connection{
		reader:   bytes.NewBufferString(""),
		writer:   bytes.NewBufferString(""),
		commands: globalCommands,
		KV:       map[string]RedisValue{},
		Config:   rawConfig,
	}
}

func processCommand(t *testing.T, previous *Connection, cmd string, expectedResponse string) *Connection {
	connection := createTestConnection(nil)
	connection.reader = bytes.NewBufferString(cmd)

	if previous != nil {
		connection.KV = previous.KV
		connection.Config = previous.Config
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
		resp("SET", "key", "value"),
		resp_ok(),
	)
	processCommand(t, connection,
		resp("GET", "key"),
		resp("value"),
	)
}

func TestKeys(t *testing.T) {
	connection := processCommand(t, nil,
		resp("KEYS", "*"),
		resp_nil(),
	)
	connection = processCommand(t, connection,
		resp("SET", "key", "value"),
		resp_ok(),
	)
	connection = processCommand(t, connection,
		resp("KEYS", "*"),
		resp_array([]string{"key"}),
	)
}

func TestConfig(t *testing.T) {
	connection := processCommand(t, nil,
		resp("SET", "key", "value"),
		resp_ok(),
	)
	connection.Config = buildConfig(strings.Split("--dir /tmp/redis-files --dbfilename dump.rdb", " "))
	processCommand(t, connection,
		resp("CONFIG", "GET", "dir"),
		resp("dir", "/tmp/redis-files"),
	)
	processCommand(t, connection,
		resp("CONFIG", "GET", "dbfilename"),
		resp("dbfilename", "dump.rdb"),
	)
}

func TestSetGetExpiry(t *testing.T) {
	connection := processCommand(t, nil,
		resp("SET", "key", "value", "px", "1000"),
		resp_ok(),
	)
	connection = processCommand(t, connection,
		resp("GET", "key"),
		resp("value"),
	)
	time.Sleep(2 * time.Second)
	connection = processCommand(t, connection,
		resp("GET", "key"),
		resp_nil(),
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

func TestDbArgs(t *testing.T) {
	cmdline := "./your_program.sh --dir /tmp/redis-files --dbfilename dump.rdb"
	config := buildConfig(strings.Split(cmdline, " "))
	if config["dir"] != "/tmp/redis-files" {
		t.Errorf("expected config[dir] = %q, got %q", "/tmp/redis-files", config["dir"])
	}
	if config["dbfilename"] != "dump.rdb" {
		t.Errorf("expected config[dbfilename] = %q, got %q", "dump.rdb", config["dbfilename"])
	}
}
