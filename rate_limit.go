package main

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// rate limit visitor
type visitor struct {
	limiter *rate.Limiter

	// store when we last saw user
	lastSeen time.Time
}

type Limiter struct {
	limit   int
	burst   int
	timeout time.Duration

	visitors map[string]*visitor
	mut      sync.RWMutex
}

func NewLimiter(limit, burst int, timeout time.Duration) *Limiter {
	return &Limiter{
		limit,
		burst,
		timeout,
		make(map[string]*visitor),
		sync.RWMutex{},
	}
}

func (self *Limiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		self.mut.Lock()
		for ip, v := range self.visitors {
			if time.Since(v.lastSeen) > self.timeout {
				delete(self.visitors, ip)
			}
		}
		self.mut.Unlock()
	}
}

func (self *Limiter) getVisitor(ip string) *rate.Limiter {

	self.mut.RLock()
	v, exists := self.visitors[ip]
	self.mut.RUnlock()

	if !exists {
		self.mut.Lock()
		limiter := rate.NewLimiter(rate.Limit(self.limit), self.burst)
		// Include the current time when creating a new visitor.
		self.visitors[ip] = &visitor{limiter, time.Now()}
		self.mut.Unlock()
		return limiter
	}

	// Update the last seen time for the visitor.
	self.mut.Lock()
	v.lastSeen = time.Now()
	self.mut.Unlock()
	return v.limiter
}

func (self *Limiter) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		limiter := self.getVisitor(ip)
		if limiter.Allow() == false {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
