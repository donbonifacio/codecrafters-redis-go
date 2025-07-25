package main

import (
	"fmt"
	"net"
	"os"
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
