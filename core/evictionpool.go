package core

import (
	"sort"
)

type PoolItem struct {
	key            string
	LastAccessedAt uint32
}

type EvictionPool struct {
	pool   []*PoolItem
	keyset map[string]*PoolItem
}

type ByIdleTime []*PoolItem

func (a ByIdleTime) Len() int {
	return len(a)
}

func (a ByIdleTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByIdleTime) Less(i, j int) bool {
	return getIdleTime(a[i].LastAccessedAt) > getIdleTime(a[j].LastAccessedAt)
}

func (pq *EvictionPool) Push(key string, lastAccessedAt uint32) {
	_, ok := pq.keyset[key]
	if ok {
		return
	}

	item := &PoolItem{key: key, LastAccessedAt: lastAccessedAt}
	if len(pq.pool) < ePoolSizeMax {
		pq.keyset[key] = item
		pq.pool = append(pq.pool, item)
		sort.Sort(ByIdleTime(pq.pool))
	} else if lastAccessedAt > pq.pool[len(pq.pool)-1].LastAccessedAt {
		pq.pool = pq.pool[1:]
		pq.keyset[key] = item
		pq.pool = append(pq.pool, item)
	}
}

func (pq *EvictionPool) Pop() *PoolItem {
	if len(pq.pool) == 0 {
		return nil
	}

	item := pq.pool[0]
	pq.pool = pq.pool[1:]
	return item
}

func newEvictionPool(size int) *EvictionPool {
	return &EvictionPool{
		pool:   make([]*PoolItem, size),
		keyset: make(map[string]*PoolItem),
	}
}

var ePoolSizeMax int = 16
var ePool *EvictionPool = newEvictionPool(0)
