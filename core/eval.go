package core

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n")
var RESP_OK []byte = []byte("+OK\r\n")
var RESP_ZERO []byte = []byte(":0\r\n")
var RESP_ONE []byte = []byte(":1\r\n")
var RESP_MINUS_1 []byte = []byte(":-1\r\n")
var RESP_MINUS_2 []byte = []byte(":-2\r\n")

func evalPING(args []string) []byte {
	var b []byte

	if len(args) >= 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'ping' command"), false)
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	return b
}

func evalBGREWRITEAOF(args []string) []byte {
	DumpAllAOF()
	return RESP_OK
}

func evalSET(args []string) []byte {

	if len(args) <= 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'set' command"), false)
	}

	var key, value string
	var exDurationMs int64 = -1
	key, value = args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return Encode(errors.New("(error) ERR syntax error"), false)
			}

			exDurationSec, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return Encode(errors.New("(error) ERR value is not an integer or out of range"), false)
			}
			exDurationMs = exDurationSec * 1000
		default:
			return Encode(errors.New("(errors) ERR syntax error"), false)
		}
	}

	Put(key, NewObj(value, exDurationMs))
	return RESP_OK
}

func evalGET(args []string) []byte {

	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'get' command"), false)
	}

	var key string = args[0]

	obj := Get(key)

	if obj == nil {
		return RESP_NIL
	}

	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		return RESP_NIL
	}
	return Encode(obj.Value, false)
}

func evalTTL(args []string) []byte {

	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'ttl' command"), false)
	}

	var key string = args[0]

	obj := Get(key)

	if obj == nil {
		return RESP_MINUS_2
	}

	if obj.ExpiresAt == -1 {
		return RESP_MINUS_1
	}
	var timeLeftForExpiration int64 = obj.ExpiresAt - time.Now().UnixMilli()

	if timeLeftForExpiration < 0 {
		return RESP_MINUS_2
	}
	return Encode(int64(timeLeftForExpiration/1000), false)
}

func evalEXPIRE(args []string) []byte {

	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'expire' command"), false)
	}
	var key string = args[0]
	obj := Get(key)
	if obj == nil {
		return RESP_ZERO
	}
	var exDurationMs int64 = -1
	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("(error) ERR value is not an integer or out of range"), false)
	}
	exDurationMs = exDurationSec * 1000
	Put(key, NewObj(obj.Value, exDurationMs))
	return RESP_ONE
}

func evalDELETE(args []string) []byte {

	var deletedKeys int64 = 0
	for i := 0; i < len(args); i++ {
		var obj = Get(args[i])
		if obj != nil {
			deletedKeys++
			Delete(args[i])
		}
	}
	return Encode(deletedKeys, false)
}

func EvalAndRespond(cmds RedisCmds, c io.ReadWriter) {
	var response []byte
	buf := bytes.NewBuffer(response)
	for _, cmd := range cmds {
		switch cmd.Cmd {
		case "PING":
			buf.Write(evalPING(cmd.Args))
		case "SET":
			buf.Write(evalSET(cmd.Args))
		case "GET":
			buf.Write(evalGET(cmd.Args))
		case "TTL":
			buf.Write(evalTTL(cmd.Args))
		case "DEL":
			buf.Write(evalDELETE(cmd.Args))
		case "EXPIRE":
			buf.Write(evalEXPIRE(cmd.Args))
		case "BGREWRITEAOF":
			buf.Write(evalBGREWRITEAOF(cmd.Args))
		default:
			buf.Write(evalPING(cmd.Args))
		}
	}
	c.Write(buf.Bytes())
}
