package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

type Connection struct {
	id       int
	conn     net.Conn
	reader   io.Reader
	writer   io.Writer
	command  string
	response string
	raw      string
	args     []string
	commands map[string]func(*Connection) error
	KV       map[string]RedisValue
	Config   map[string]string
}

type RedisValue struct {
	Value     string
	Type      string
	ExpiresAt *time.Time
}

func handleConnection(connection *Connection) error {
	fmt.Printf("[%d] Accepted connection for client\n", connection.id)
	for true {
		parts, err := readRedisCmd(connection.reader)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Printf("[%d] EOF\n", connection.id)
				return nil
			}
			fmt.Println("Error reading from connection:", err.Error())
			os.Exit(1)
		}
		connection.raw = strings.Join(parts, " ")
		connection.command = strings.ToLower(parts[0])
		connection.args = parts[1:]

		fmt.Printf("[%d] Cmd: '%s'\n", connection.id, connection.raw)
		if command, ok := connection.commands[connection.command]; ok {
			err := command(connection)
			if err != nil {
				fmt.Printf("[%d] Error executing command: %s\n", connection.id, err.Error())
				connection.response = "-ERR " + err.Error()
			}
		} else {
			fmt.Printf("[%d] Unknown command '%s'\n", connection.id, connection.raw)
			connection.response = "-ERR unknown command '" + connection.command + "'\r\n"
		}
		fmt.Printf("[%d] Return: '%s'\n", connection.id, connection.response)
		connection.writer.Write([]byte(connection.response))
	}
	return nil
}
