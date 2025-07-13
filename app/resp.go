package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

func resp(parts ...string) string {
	if len(parts) == 0 {
		return resp_nil()
	}

	var sb strings.Builder
	if len(parts) > 1 {
		sb.WriteString(fmt.Sprintf("*%d\r\n", len(parts)))
	}
	for _, part := range parts {
		sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(part), part))
	}
	return sb.String()
}

func resp_ok() string {
	return resp_success("OK")
}

func resp_nil() string {
	return "$-1\r\n"
}

func resp_error(err error) string {
	return fmt.Sprintf("-ERR %s\r\n", err.Error())
}

func resp_success(msg string) string {
	return fmt.Sprintf("+%s\r\n", msg)
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
