package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rlBucket struct {
	windowStart time.Time
	count       int
}

// RateLimit is a minimal in-memory limiter (dev-grade).
// For prod, replace with Redis/distributed limiter.
func RateLimit(perMin int) gin.HandlerFunc {
	if perMin <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	var (
		mu      sync.Mutex
		buckets = map[string]*rlBucket{}
	)
	return func(c *gin.Context) {
		key := c.ClientIP() + "|" + c.FullPath()
		now := time.Now()

		mu.Lock()
		b, ok := buckets[key]
		if !ok {
			b = &rlBucket{windowStart: now}
			buckets[key] = b
		}
		if now.Sub(b.windowStart) >= time.Minute {
			b.windowStart = now
			b.count = 0
		}
		b.count++
		n := b.count
		mu.Unlock()

		if n > perMin {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
			return
		}
		c.Next()
	}
}
