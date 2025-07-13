package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	id := 0
	for true {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		id += 1

		connection := Connection{
			id:       id,
			conn:     conn,
			reader:   conn,
			writer:   conn,
			commands: globalCommands,
			KV:       map[string]RedisValue{},
			Config:   buildConfig(os.Args),
		}

		go func(conn *Connection) {
			defer connection.conn.Close()
			handleConnection(conn)
		}(&connection)
	}
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

		fmt.Printf("[%d] Received: '%s'\n", connection.id, connection.raw)
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
		connection.writer.Write([]byte(connection.response))
	}
	return nil
}

func buildConfig(args []string) map[string]string {
	config := make(map[string]string)
	for i, arg := range args {
		if arg == "--dir" {
			config["dir"] = args[i+1]
		} else if arg == "--dbfilename" {
			config["dbfilename"] = args[i+1]
		}
	}
	return config
}
