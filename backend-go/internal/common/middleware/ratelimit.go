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
	var mu sync.Mutex
	buckets := map[string]bucket{}
	return func(c *gin.Context) {
		if perMin <= 0 {
			c.Next()
			return
		}
		key := c.ClientIP() + c.FullPath()
		now := time.Now()
		mu.Lock()
		b := buckets[key]
		if now.After(b.reset) {
			b = bucket{reset: now.Add(time.Minute)}
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
