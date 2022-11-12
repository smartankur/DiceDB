package core

import "github.com/smartankur/dice/config"

func evictFirst() {
	for k := range store {
		delete(store, k)
		return
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
	}

}
