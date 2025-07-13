package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

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
	KV       map[string]string
}

var globalCommands = map[string]func(*Connection) error{
	"ping": ping,
	"echo": echo,
	"set":  set,
	"get":  get,
}

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
			KV:       map[string]string{},
		}

		go func(conn *Connection) {
			defer connection.conn.Close()
			handleConnection(conn)
		}(&connection)
	}
}

func ping(c *Connection) error {
	c.response = "+PONG\r\n"
	return nil
}

func set(c *Connection) error {
	if len(c.args) != 2 {
		return fmt.Errorf("-ERR wrong number of arguments for 'set' command: %s\r\n", c.raw)
	}
	c.KV[c.args[0]] = c.args[1]
	c.response = "+OK\r\n"
	return nil
}

func get(c *Connection) error {
	if len(c.args) != 1 {
		return fmt.Errorf("-ERR wrong number of arguments for 'get' command: %s\r\n", c.raw)
	}
	value := c.KV[c.args[0]]
	c.response = fmt.Sprintf("$%v\r\n%v\r\n", len(value), value)
	return nil
}

func echo(c *Connection) error {
	c.response = fmt.Sprintf("+%s\r\n", strings.Join(c.args, " "))
	return nil
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

func readLine(conn io.Reader) (string, error) {
	var line []byte
	buf := make([]byte, 1)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return "", err
		}
		if n > 0 {
			line = append(line, buf[0])
			if buf[0] == '\n' {
				break
			}
		}
	}
	return string(line), nil
}

func readRedisCmd(conn io.Reader) ([]string, error) {
	parts := []string{}

	// Read first line: *<number_of_elements>\r\n
	firstLine, err := readLine(conn)
	if err != nil {
		return nil, err
	}
	firstLine = strings.TrimSpace(firstLine)
	if !strings.HasPrefix(firstLine, "*") {
		return nil, fmt.Errorf("expected '*' at start of command")
	}
	numElements, err := strconv.Atoi(firstLine[1:])
	if err != nil {
		return nil, fmt.Errorf("invalid number of elements: %v", err)
	}

	for i := 0; i < numElements; i++ {
		// Read $<length>\r\n
		lenLine, err := readLine(conn)
		if err != nil {
			return nil, err
		}
		lenLine = strings.TrimSpace(lenLine)
		if !strings.HasPrefix(lenLine, "$") {
			return nil, fmt.Errorf("expected '$' at start of element")
		}
		_, err = strconv.Atoi(lenLine[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid element length: %v", err)
		}

		// Read <word>\r\n
		word, err := readLine(conn)
		if err != nil {
			return nil, err
		}
		parts = append(parts, strings.TrimSpace(word))
	}

	return parts, nil
}
