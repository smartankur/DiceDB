package core

type Obj struct {
	TypeEncoding   uint8
	Value          interface{} // same as pointers in c++
	LastAccessedAt uint32
}

var OBJ_TYPE_STRING uint8 = 0 << 4

var OBJ_ENCODING_RAW uint8 = 0
var OBJ_ENCODING_INT uint8 = 1
var OBJ_ENCODING_EMBSTR uint8 = 8
