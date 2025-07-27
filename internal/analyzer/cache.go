package analyzer

import (
	"sync"
	"time"
)

type cacheEntry struct {
	Result    *Result
	Timestamp time.Time
}

var (
	cache     = make(map[string]cacheEntry)
	cacheLock sync.RWMutex
	cacheTTL  = 10 * time.Minute
)

func GetFromCache(url string) (*Result, bool) {
	cacheLock.RLock()
	defer cacheLock.RUnlock()
	entry, ok := cache[url]
	if !ok || time.Since(entry.Timestamp) > cacheTTL {
		return nil, false
	}
	return entry.Result, true
}

func StoreInCache(url string, res *Result) {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	cache[url] = cacheEntry{
		Result:    res,
		Timestamp: time.Now(),
	}
}
