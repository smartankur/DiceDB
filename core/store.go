package core

import (
	"time"

	"github.com/smartankur/dice/config"
)

var store map[string]*Obj
var expires map[*Obj]uint64

func init() {
	store = make(map[string]*Obj)
	expires = make(map[*Obj]uint64)
}

func setExpiry(obj *Obj, exDuarationMs int64) {
	expires[obj] = uint64(time.Now().UnixMilli()) + uint64(exDuarationMs)
}

func NewObj(value interface{}, durationMs int64, oType uint8, oEnc uint8) *Obj {
	obj := &Obj{
		Value:          value,
		TypeEncoding:   oType | oEnc,
		LastAccessedAt: getCurrentClock(),
	}

	if durationMs > 0 {
		setExpiry(obj, durationMs)
	}
	return obj
}

func getCurrentClock() uint32 {
	return uint32(time.Now().UnixMilli()) & 0x00FFFFFF
}

func Put(k string, obj *Obj) {
	if len(store) >= config.KeysLimit {
		evict()
	}
	obj.LastAccessedAt = getCurrentClock()
	store[k] = obj
	if KeySpaceStat[0] == nil {
		KeySpaceStat[0] = make(map[string]int)
	}
	KeySpaceStat[0]["keys"]++
}

func Get(k string) *Obj {
	v := store[k]
	if v != nil {
		if hasExpired(v) {
			Delete(k)
			return nil
		}
	}
	v.LastAccessedAt = getCurrentClock()
	return v
}

func hasExpired(obj *Obj) bool {
	exp, ok := expires[obj]
	if !ok {
		return false
	}
	return exp <= uint64(time.Now().UnixMilli())
}

func getExpiry(obj *Obj) (uint64, bool) {
	exp, ok := expires[obj]
	return exp, ok
}

func Delete(k string) bool {
	if obj, ok := store[k]; ok {
		delete(store, k)
		delete(expires, obj)
		KeySpaceStat[0]["keys"]--
		return true
	}
	return false
}
