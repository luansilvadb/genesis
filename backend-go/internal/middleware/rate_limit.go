package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		now := time.Now()
		for ip, timestamps := range rl.visitors {
			var valid []time.Time
			for _, ts := range timestamps {
				if now.Sub(ts) < rl.window {
					valid = append(valid, ts)
				}
			}
			if len(valid) == 0 {
				delete(rl.visitors, ip)
			} else {
				rl.visitors[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Prevent unbounded map growth under high traffic.
	const maxVisitors = 10000
	if len(rl.visitors) >= maxVisitors {
		// Aggressively evict: clear all stale entries.
		now := time.Now()
		for ipKey, timestamps := range rl.visitors {
			var valid []time.Time
			for _, ts := range timestamps {
				if now.Sub(ts) < rl.window {
					valid = append(valid, ts)
				}
			}
			if len(valid) == 0 {
				delete(rl.visitors, ipKey)
			} else {
				rl.visitors[ipKey] = valid
			}
		}
		// If still full, reject the request to protect memory.
		if len(rl.visitors) >= maxVisitors {
			return false
		}
	}

	now := time.Now()
	timestamps := rl.visitors[ip]

	var valid []time.Time
	for _, ts := range timestamps {
		if now.Sub(ts) < rl.window {
			valid = append(valid, ts)
		}
	}

	if len(valid) >= rl.limit {
		rl.visitors[ip] = valid
		return false
	}

	rl.visitors[ip] = append(valid, now)
	return true
}

// RateLimit returns a per-instance rate limiter. When deployed behind a load
// balancer with N instances, the effective limit is limit × N. For production
// deployments requiring strict rate limiting, use a distributed store (Redis).
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	rl := newRateLimiter(limit, window)
	return func(c *gin.Context) {
		// Use RemoteIP() instead of ClientIP() to ignore X-Forwarded-For /
		// X-Real-IP headers, preventing trivial rate-limit bypass via header
		// spoofing. Strip the TCP port so all connections from the same IP
		// share the same rate-limit bucket.
		ip := c.Request.RemoteAddr
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		if !rl.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "Muitas requisições. Aguarde um momento e tente novamente.",
			})
			return
		}
		c.Next()
	}
}
