package cache

import (
	"time"

	"github.com/kiremitrov123/onboarding/stockprice/model"
	"github.com/patrickmn/go-cache"
)

const cachingInProgress = "__caching__"

type LocalCache struct {
	c *cache.Cache
}

func NewLocalCache(defaultExpiration, cleanupInterval time.Duration) *LocalCache {
	return &LocalCache{
		c: cache.New(defaultExpiration, cleanupInterval),
	}
}

func (lc *LocalCache) Get(symbol string) (*model.Price, bool) {
	val, found := lc.c.Get(symbol)
	if !found || val == cachingInProgress {
		return nil, false
	}

	price, ok := val.(model.Price)
	if !ok {
		return nil, false
	}
	return &price, true
}

func (lc *LocalCache) Set(symbol string, price model.Price) {
	val, found := lc.c.Get(symbol)

	// Only allow setting if key is still a placeholder or doesn't exist
	if found && val != cachingInProgress {
		return
	}

	lc.c.Set(symbol, price, cache.DefaultExpiration)
}

func (lc *LocalCache) SetPlaceholder(symbol string) {
	lc.c.Set(symbol, cachingInProgress, time.Minute)
}

func (lc *LocalCache) Invalidate(symbol string) {
	lc.c.Delete(symbol)
}
