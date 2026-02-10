package globals

import (
	"container/list"
	"sync"
)

type LRUCache[K comparable, V any] struct {
	capacity int
	cache    map[K]*list.Element
	list     *list.List
	mu       sync.Mutex
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]*list.Element),
		list:     list.New(),
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.cache[key]; ok {
		c.list.MoveToFront(element)
		return element.Value.(*entry[K, V]).value, true
	}

	var zero V
	return zero, false
}

func (c *LRUCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.cache[key]; ok {
		c.list.MoveToFront(element)
		element.Value.(*entry[K, V]).value = value
		return
	}

	if c.list.Len() >= c.capacity {
		last := c.list.Back()
		if last != nil {
			c.list.Remove(last)
			delete(c.cache, last.Value.(*entry[K, V]).key)
		}
	}

	newEntry := &entry[K, V]{key: key, value: value}
	element := c.list.PushFront(newEntry)
	c.cache[key] = element
}

func (c *LRUCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.cache[key]; ok {
		c.list.Remove(element)
		delete(c.cache, key)
	}
}
