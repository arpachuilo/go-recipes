package main

import (
	"crypto/sha1"
	"fmt"
	"sync"
)

type ETags[T comparable] struct {
	etags map[T]string
	ids   map[string]T

	mut sync.RWMutex
}

func (self *ETags[T]) HasID(id T) bool {
	self.mut.RLock()
	_, ok := self.etags[id]
	self.mut.RUnlock()
	return ok
}

func (self *ETags[T]) HasETag(etag string) bool {
	self.mut.RLock()
	_, ok := self.ids[etag]
	self.mut.RUnlock()
	return ok
}

func (self *ETags[T]) InvalidateByETag(etag string) {
	if id, ok := self.ids[etag]; ok {
		self.mut.Lock()
		delete(self.ids, etag)
		delete(self.etags, id)
		self.mut.Unlock()
	}
}

func (self *ETags[T]) InvalidateByID(id T) {
	if etag, ok := self.etags[id]; ok {
		self.mut.Lock()
		delete(self.ids, etag)
		delete(self.etags, id)
		self.mut.Unlock()
	}
}

func (self *ETags[T]) Add(id T, body []byte, weak bool) string {
	hash := sha1.Sum(body)
	etag := fmt.Sprintf("\"%d-%x\"", int(len(hash)), hash)

	if weak {
		etag = "W/" + etag
	}

	self.mut.Lock()
	self.etags[id] = etag
	self.ids[etag] = id
	self.mut.Unlock()
	return etag
}

func NewETags[T comparable]() *ETags[T] {
	return &ETags[T]{
		etags: make(map[T]string),
		ids:   make(map[string]T),
		mut:   sync.RWMutex{},
	}
}
