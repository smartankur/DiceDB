package core

import (
	"errors"
)

func readLength(data []byte) (int, int) {
	pos, length := 0, 0
	for pos = range data {
		b := data[pos]
		if b < '0' || b > '9' {
			return length, pos + 2
		}
		length = length + int(b-'0')
	}
	return 0, 0
}

func parseSimpleString(data []byte) (string, int, error) {
	pos := 1
	for ; data[pos] != '\r'; pos++ {
	}
	return string(data[1:pos]), pos + 2, nil
}

func parseError(data []byte) (string, int, error) {
	return parseSimpleString(data)
}

func parseInt64(data []byte) (int64, int, error) {
	pos := 1
	var value int64 = 0
	for ; data[pos] != '\r'; pos++ {
		value = value*10 + int64(data[pos]-'0')
	}
	return value, pos + 2, nil
}

func parseBulkString(data []byte) (string, int, error) {
	pos := 1
	len, delta := readLength(data[pos:])
	pos += delta
	return string(data[pos:(pos + len)]), pos + len + 2, nil
}

func parseArray(data []byte) (interface{}, int, error) {
	pos := 1
	count, delta := readLength(data[pos:])
	pos += delta
	var elems []interface{} = make([]interface{}, count)
	for i := range elems {
		elem, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		elems[i] = elem
		pos += delta
	}
	return elems, pos, nil
}

func DecodeOne(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("No data")
	}
	switch data[0] {
	case '+':
		return parseSimpleString(data)
	case '-':
		return parseError(data)
	case '*':
		return parseArray(data)
	case '$':
		return parseBulkString(data)
	case ':':
		return parseInt64(data)
	}
	return nil, 0, nil
}

func Decode(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, errors.New("No data")
	}
	value, _, err := DecodeOne(data)
	return value, err
}
