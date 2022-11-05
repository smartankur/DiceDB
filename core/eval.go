package core

import (
	"errors"
	"io"
	"log"
	"strconv"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n")

func evalPING(args []string, c io.ReadWriter) error {
	var b []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := c.Write(b)
	return err
}

func evalSET(args []string, c io.ReadWriter) error {

	if len(args) <= 1 {
		return errors.New("ERR wrong number of arguments for 'set' command")
	}

	var key, value string
	var exDurationMs int64 = -1
	key, value = args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return errors.New("(error) ERR syntax error")
			}

			exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return errors.New("(error) ERR value is not an integer or out of range")
			}
			exDurationMs = exDurationSec * 1000
		default:
			return errors.New("(errors) ERR syntax error")
		}
	}

	Put(key, NewObj(value, exDurationMs))
	c.Write([]byte("+OK\r\n"))
	return nil
}

func evalGET(args []string, c io.ReadWriter) error {

	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'get' command")
	}

	var key string = args[0]

	obj := Get(key)

	if obj == nil {
		c.Write(RESP_NIL)
		return nil
	}

	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		c.Write(RESP_NIL)
		return nil
	}

	c.Write(Encode(obj.Value, false))
	return nil
}

func evalTTL(args []string, c io.ReadWriter) error {

	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'ttl' command")
	}

	var key string = args[0]

	obj := Get(key)

	if obj == nil {
		c.Write([]byte(":-2\r\n"))
		return nil
	}

	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		c.Write([]byte(":-1\r\n"))
		return nil
	}
	var timeLeftForExpiration int64 = obj.ExpiresAt - time.Now().UnixMilli()

	if timeLeftForExpiration < 0 {
		c.Write([]byte(":-2\r\n"))
		return nil
	}

	c.Write(Encode(int64(timeLeftForExpiration/1000), false))
	return nil
}

func EvalAndRespond(cmd *RedisCmd, c io.ReadWriter) error {
	log.Println("comamnd:", cmd.Cmd)
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args, c)
	case "SET":
		return evalSET(cmd.Args, c)
	case "GET":
		return evalGET(cmd.Args, c)
	case "TTL":
		return evalTTL(cmd.Args, c)
	default:
		return evalPING(cmd.Args, c)
	}
}
