package main

import "fyne.io/fyne/v2/canvas"

type LRUCache struct {
	capacity int
	cache    map[string]*CacheNode
	head     *CacheNode
	tail     *CacheNode
}

func (c *LRUCache) moveToFront(node *CacheNode) {
	if node == c.head {
		return
	}
	if node == c.tail {
		c.tail = node.prev
		c.tail.next = nil
	} else if node.prev != nil {
		node.prev.next = node.next
		node.next.prev = node.prev
	}
	node.prev = nil
	node.next = c.head
	if c.head != nil {
		c.head.prev = node
	}
	c.head = node
}

func (c *LRUCache) add(key string, image *canvas.Image) {
	if node, exists := c.cache[key]; exists {
		node.image = image
		c.moveToFront(node)
		return
	}

	node := &CacheNode{key, image, nil, nil}
	c.cache[key] = node

	if c.head == nil {
		c.head = node
		c.tail = node
	} else {
		node.next = c.head
		c.head.prev = node
		c.head = node
	}

	if len(c.cache) > c.capacity {
		delete(c.cache, c.tail.key)
		c.tail = c.tail.prev
		if c.tail != nil {
			c.tail.next = nil
		}
	}
}

func (c *LRUCache) get(key string) (*canvas.Image, bool) {
	if node, exists := c.cache[key]; exists {
		c.moveToFront(node)
		return node.image, true
	}
	return nil, false
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*CacheNode),
	}
}
