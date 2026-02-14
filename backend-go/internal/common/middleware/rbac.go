package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequirePermission(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("permissions")
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		perms, ok := v.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		for _, p := range perms {
			if p == name {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}
