package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	count int
	reset time.Time
}

func RateLimit(perMin int) gin.HandlerFunc {
	if perMin <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	var mu sync.Mutex
	buckets := map[string]bucket{}

	return func(c *gin.Context) {
		key := c.ClientIP() + ":" + c.FullPath()
		now := time.Now()

		mu.Lock()
		b := buckets[key]
		if b.reset.IsZero() || now.After(b.reset) {
			b = bucket{reset: now.Add(60 * time.Second)}
		}
		b.count++
		buckets[key] = b
		mu.Unlock()

		if b.count > perMin {
			c.AbortWithStatusJSON(429, gin.H{"error": "rate limit"})
			return
		}

		c.Next()
	}
}
