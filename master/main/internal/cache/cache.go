package cache

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	rwMux             sync.RWMutex
	items             map[string]Item
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	countGlobal       int
}

type Item struct {
	Value      map[string]string
	Count      int
	Expiration int64
	Active     bool
}

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {

	items := make(map[string]Item)

	cache := Cache{
		items:             items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		countGlobal:       0,
	}

	if cleanupInterval > 0 {
		cache.StartGC()
	}

	return &cache
}

func (c *Cache) Set(key string, isActive bool, value map[string]string) {
	expiration := time.Now().Add(c.defaultExpiration).UnixNano()
	c.rwMux.Lock()

	c.countGlobal++

	item, ok := c.items[key]
	if !ok {
		c.items[key] = Item{
			Value:      value,
			Expiration: expiration,
			Active:     isActive,
			Count:      1,
		}
	}
	item.Value = value
	c.rwMux.Unlock()
	fmt.Println(len(c.items))
	if len(c.items) >= 20 {
		fmt.Println("тут")
		if keys := c.expiredKeys(); len(keys) != 0 {
			fmt.Println("тут 1")
			c.clearItems(keys)
		}
	}
}

func (c *Cache) Get(key string) (map[string]string, bool, bool) {

	c.rwMux.RLock()

	defer c.rwMux.RUnlock()

	item, found := c.items[key]

	if !found {
		return nil, false, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false, false
		}

	}

	item.Count++
	c.countGlobal++

	return item.Value, item.Active, true
}

func (c *Cache) Delete(keys []string) {

	c.rwMux.Lock()

	defer c.rwMux.Unlock()

	for _, key := range keys {
		if _, found := c.items[key]; !found {
			return
		}

		delete(c.items, key)
	}
}

func (c *Cache) StartGC() {
	go c.GC()
}

func (c *Cache) GC() {

	for {
		select {
		case <-time.After(c.cleanupInterval):
			if c.items == nil {
				return
			}
			if keys := c.expiredKeys(); len(keys) != 0 {
				c.clearItems(keys)
			}
		}
	}
}

func (c *Cache) expiredKeys() (keys []string) {
	fmt.Println("тут 4")
	c.rwMux.RLock()

	defer c.rwMux.RUnlock()

	for k, i := range c.items {
		if (time.Now().UnixNano() > i.Expiration && i.Expiration > 0) || !check(i.Count, c.countGlobal) {
			keys = append(keys, k)
		}
		i.Count = 0
	}
	c.countGlobal = 0

	return
}

func (c *Cache) clearItems(keys []string) {

	c.rwMux.Lock()

	defer c.rwMux.Unlock()

	for _, k := range keys {
		fmt.Println(k)
		delete(c.items, k)
	}
}

func check(count int, globalCount int) bool {
	checkCount := float64(count) / float64(globalCount)
	return checkCount >= 0.2
}
