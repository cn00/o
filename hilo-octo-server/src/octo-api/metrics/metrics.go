package metrics

import (
	"expvar"
)

var (
	Map = expvar.NewMap("cache")

	MemoryCacheHit  = new(expvar.Int)
	MemoryCacheMiss = new(expvar.Int)
)

func init() {
	Map.Set("memory_cache_hit", MemoryCacheHit)
	Map.Set("memory_cache_miss", MemoryCacheMiss)
}
