package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var globalCommands = map[string]func(*Connection) error{
	"ping":   ping,
	"echo":   echo,
	"set":    set,
	"get":    get,
	"keys":   keys,
	"config": handleConfig,
}

func ping(c *Connection) error {
	c.response = resp_success("PONG")
	return nil
}

func set(c *Connection) error {
	if len(c.args) < 2 {
		return fmt.Errorf("-ERR wrong number of arguments for 'set' command: %s\r\n", c.raw)
	}
	value := RedisValue{
		Value:     c.args[1],
		Type:      "string",
		ExpiresAt: nil,
	}
	if len(c.args) == 4 {
		duration, err := strconv.Atoi(c.args[3])
		if err != nil {
			return fmt.Errorf("-ERR invalid duration: %s\r\n", err.Error())
		}
		expires_at := time.Now().UTC().Add(time.Duration(duration) * time.Millisecond)
		value.ExpiresAt = &expires_at
	}
	c.KV[c.args[0]] = value
	c.response = resp_ok()
	return nil
}

func get(c *Connection) error {
	if len(c.args) != 1 {
		return fmt.Errorf("-ERR wrong number of arguments for 'get' command: %s\r\n", c.raw)
	}

	value, err := raw_get(c, c.args[0])
	if err != nil {
		return err
	}
	c.response = value
	return nil
}

func raw_get(c *Connection, key string) (string, error) {
	value := c.KV[key]

	now := time.Now().UTC()
	if value.ExpiresAt != nil && now.After(*value.ExpiresAt) {
		delete(c.KV, key)
		return "", nil
	}

	return resp(value.Value), nil
}

func echo(c *Connection) error {
	c.response = fmt.Sprintf("+%s\r\n", strings.Join(c.args, " "))
	return nil
}

func keys(c *Connection) error {
	if len(c.args) != 1 {
		return fmt.Errorf("-ERR wrong number of arguments for 'keys' command: %s\r\n", c.raw)
	}

	var results []string
	for k := range c.KV {
		val, _ := raw_get(c, k)
		if len(val) > 0 {
			results = append(results, k)
		}
	}

	c.response = resp_array(results)
	return nil
}

func handleConfig(c *Connection) error {
	key := c.args[1]
	value := c.Config[key]

	c.response = resp(key, value)
	return nil
}
