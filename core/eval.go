package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n")
var RESP_OK []byte = []byte("+OK\r\n")
var RESP_QUEUED []byte = []byte("+QUEUED\r\n")
var RESP_ZERO []byte = []byte(":0\r\n")
var RESP_ONE []byte = []byte(":1\r\n")
var RESP_MINUS_1 []byte = []byte(":-1\r\n")
var RESP_MINUS_2 []byte = []byte(":-2\r\n")

var txnCommands map[string]bool

func init() {
	txnCommands = map[string]bool{"EXEC": true, "DISCARD": true}
}

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
	oType, oEnc := deduceTypeEncoding(value)

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return Encode(errors.New("ERR syntax error"), false)
			}

			exDurationSec, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return Encode(errors.New("ERR value is not an integer or out of range"), false)
			}
			exDurationMs = exDurationSec * 1000
		default:
			return Encode(errors.New("ERR syntax error"), false)
		}
	}

	Put(key, NewObj(value, exDurationMs, oType, oEnc))
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

	if hasExpired(obj) {
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

	exp, isExpirySet := getExpiry(obj)
	if !isExpirySet {
		return RESP_MINUS_1
	}

	if uint64(time.Now().UnixMilli()) > exp {
		return RESP_MINUS_2
	}

	durationMs := exp - uint64(time.Now().UnixMilli())
	return Encode(int64(durationMs/1000), false)
}

func evalEXPIRE(args []string) []byte {

	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'expire' command"), false)
	}

	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("(error) ERR value is not an integer or out of range"), false)
	}

	var key string = args[0]
	obj := Get(key)
	if obj == nil {
		return RESP_ZERO
	}

	setExpiry(obj, exDurationSec*1000)
	return RESP_ONE
}

func evalINCR(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments"), false)
	}

	var key string = args[0]
	obj := Get(key)
	if obj == nil {
		obj = NewObj("0", -1, OBJ_TYPE_STRING, OBJ_ENCODING_INT)
		Put(key, obj)
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_STRING); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_INT); err != nil {
		return Encode(err, false)
	}

	i, _ := strconv.ParseInt(obj.Value.(string), 10, 64)
	i++
	obj.Value = strconv.FormatInt(i, 10)
	log.Println("Incremented value   ", i)
	return Encode(i, false)
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

func evalMULTI(args []string) []byte {
	return RESP_OK
}

func execteCommand(cmd *RedisCmd, c *Client) []byte {
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args)
	case "SET":
		return evalSET(cmd.Args)
	case "GET":
		return evalGET(cmd.Args)
	case "TTL":
		return evalTTL(cmd.Args)
	case "DEL":
		return evalDELETE(cmd.Args)
	case "EXPIRE":
		return evalEXPIRE(cmd.Args)
	case "BGREWRITEAOF":
		return evalBGREWRITEAOF(cmd.Args)
	case "INCR":
		return evalINCR(cmd.Args)
	case "INFO":
		return evalINFO(cmd.Args)
	case "MULTI":
		c.TxnBegin()
		return evalMULTI(cmd.Args)
	case "EXEC":
		if !c.isTxn {
			return Encode(errors.New("ERR EXEC without MULTI"), false)
		}
		return c.TxnExec()
	case "DISCARD":
		if !c.isTxn {
			return Encode(errors.New("ERR DISCARD without MULTI"), false)
		}
		c.TxnDiscard()
		return RESP_OK
	default:
		return evalPING(cmd.Args)

	}
}

func executeCommandToBuffer(cmd *RedisCmd, buf *bytes.Buffer, c *Client) {
	buf.Write(execteCommand(cmd, c))
}

func EvalAndRespond(cmds RedisCmds, c *Client) {
	var response []byte
	buf := bytes.NewBuffer(response)
	for _, cmd := range cmds {

		if !c.isTxn {
			executeCommandToBuffer(cmd, buf, c)
			continue
		}

		if !txnCommands[cmd.Cmd] {
			c.TxnQueue(cmd)
			buf.Write(RESP_QUEUED)
		} else {
			executeCommandToBuffer(cmd, buf, c)
		}
	}
	c.Write(buf.Bytes())
}

func evalINFO(args []string) []byte {
	var info []byte
	buf := bytes.NewBuffer(info)
	buf.WriteString("# KeySpace \r\n")
	for i := range KeySpaceStat {
		buf.WriteString(fmt.Sprintf("db%d:keys=%d,epxires=0,avg_ttl=0\r\n", i, KeySpaceStat[i]["keys"]))
	}
	return Encode(buf.String(), false)
}
