package core

import "github.com/smartankur/dice/config"

func evictFirst() {
	for k := range store {
		delete(store, k)
		return
	}
}

func getIdleTime(LastAccessedAt uint32) uint32 {
	c := getCurrentClock()
	if c >= LastAccessedAt {
		return c - LastAccessedAt
	}

	return (0x00FFFFFF - LastAccessedAt) + c
}

func populateEvictionPool() {
	sampleSize := 5
	for k := range store {
		ePool.Push(k, store[k].LastAccessedAt)
		sampleSize--
		if sampleSize == 0 {
			break
		}
	}
}

func evictAllKeysLRU() {
	populateEvictionPool()
	evictionCount := int16(config.EvictionRatio * float64(config.KeysLimit))

	for i := 0; i < int(evictionCount) && len(ePool.pool) > 0; i++ {
		item := ePool.Pop()
		if item == nil {
			return
		}
		Delete(item.key)
	}
}

func evictAllKeysAtRandom() {
	evictCount := int64(config.EvictionRatio * float64(config.KeysLimit))

	for k := range store {
		Delete(k)
		evictCount--
		if evictCount <= 0 {
			break
		}
	}
}

func evict() {
	switch config.EvictionStrategy {
	case "simple-first":
		evictFirst()
	case "allkeys-random":
		evictAllKeysAtRandom()
	case "allkey-lru":
		evictAllKeysLRU()
	}

}
