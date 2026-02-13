package middleware

import "github.com/gin-gonic/gin"

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("permissions")
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}
		for _, p := range v.([]string) {
			if p == permission {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}
