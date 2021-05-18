package cache

import lru "github.com/hashicorp/golang-lru"

const (
	listCacheSize = 100
	urlCacheSize  = 1024 * 1024
)

var (
	listCache *lru.Cache
	urlCache  *lru.Cache
)

func Setup() {
	listCache = newLRUCache(listCacheSize)
	urlCache = newLRUCache(urlCacheSize)
}

func newLRUCache(size int) *lru.Cache {
	c, err := lru.New(size)
	if err != nil {
		panic(err)
	}
	return c
}
