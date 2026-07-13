package internal

import (
	"sync"
	"time"
)

type Cache struct {
	Elements map[string]cacheEntry
	Mutex    sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) *Cache {
	newCache := &Cache{
		Elements: map[string]cacheEntry{},
	}

	go newCache.reapLoop(interval)

	return newCache
}

func (c *Cache) Add(key string, val []byte) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if c.Elements == nil {
		c.Elements = make(map[string]cacheEntry)
	}

	c.Elements[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	var zero []byte
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if val, ok := c.Elements[key]; ok {
		return val.val, true
	} else {
		return zero, false
	}
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for t := range ticker.C {
		c.Mutex.Lock()
		for key, value := range c.Elements {
			if t.Sub(value.createdAt) > interval {
				delete(c.Elements, key)
			}
		}
		c.Mutex.Unlock()
	}
}
