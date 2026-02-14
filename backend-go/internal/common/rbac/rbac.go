package rbac

import "github.com/gin-gonic/gin"

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms, ok := c.Get("permissions")
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}
		for _, p := range perms.([]string) {
			if p == permission {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}
