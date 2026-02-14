package middleware

import "github.com/gin-gonic/gin"

func RequirePermission(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("permissions")
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}

		permissions, ok := v.([]string)
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}

		for _, p := range permissions {
			if p == name {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}
