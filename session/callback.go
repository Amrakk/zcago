package session

import (
	"sync"
	"time"
)

const defaultTTL = 5 * time.Minute

type UploadEventData struct {
	FileURL string `json:"fileUrl"`
	FileID  string `json:"fileId"`
}

type UploadCallback = func(data UploadEventData) any

type CallbacksMap struct {
	mu     sync.Mutex
	items  map[string]UploadCallback
	timers map[string]*time.Timer
}

func NewCallbacksMap() *CallbacksMap {
	return &CallbacksMap{
		items:  make(map[string]UploadCallback),
		timers: make(map[string]*time.Timer),
	}
}

func (c *CallbacksMap) Set(key string, value UploadCallback, ttl time.Duration) *CallbacksMap {
	if ttl <= 0 {
		ttl = defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if t, ok := c.timers[key]; ok {
		t.Stop()
	}
	c.items[key] = value

	c.timers[key] = time.AfterFunc(ttl, func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.items, key)
		delete(c.timers, key)
	})

	return c
}

func (c *CallbacksMap) Get(key string) (UploadCallback, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cb, ok := c.items[key]
	return cb, ok
}

func (c *CallbacksMap) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	if t, ok := c.timers[key]; ok {
		t.Stop()
		delete(c.timers, key)
	}
}

func (c *CallbacksMap) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

func (c *CallbacksMap) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, t := range c.timers {
		t.Stop()
		delete(c.timers, k)
	}
	clear(c.items)
}
