package main

import (
	"io"
	"net"
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
