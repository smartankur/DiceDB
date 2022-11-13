package core

import (
	"log"
)

func expireSample() float32 {
	var limit int = 20
	var expiredCount int = 0

	for key, obj := range store {
		limit--
		if hasExpired(obj) {
			Delete(key)
			expiredCount++
		}

		if limit == 0 {
			break
		}
	}
	return float32(expiredCount) / float32(20.0)
}

func DeleteExpiredKeys() {
	for {
		frac := expireSample()

		if frac < 0.25 {
			break
		}
	}
	log.Println("deleted the expired but undeleted keys, total keys", len(store))
}
