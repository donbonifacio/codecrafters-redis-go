package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
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
		go handleConnection(conn, id)
	}
}

func handleConnection(conn net.Conn, id int) {
	defer conn.Close()
	fmt.Printf("[%d] Accepted connection for client\n", id)
	for true {
		line, err := readLine(conn)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Printf("[%d] EOF\n", id)
				return
			}
			fmt.Println("Error reading from connection:", err.Error())
			os.Exit(1)
		}
		line = strings.TrimSpace(line)

		fmt.Printf("[%d] Received: '%s'\n", id, line)
		if line == "PING" {
			conn.Write([]byte("+PONG\r\n"))
		}
	}
}

// readLine reads bytes from the connection until a newline is encountered.
func readLine(conn net.Conn) (string, error) {
	var buf []byte
	tmp := make([]byte, 1)
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			return "", err
		}
		if n > 0 {
			buf = append(buf, tmp[0])
			if tmp[0] == '\n' {
				break
			}
		}
	}
	return string(buf), nil

}
