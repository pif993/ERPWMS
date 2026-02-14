package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID sets/propagates a request id.
// It sets:
// - gin key: "request_id"
// - request context value: "request_id" (so ctx.Value("request_id") works)
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set("request_id", rid)
		c.Writer.Header().Set("X-Request-Id", rid)

		ctx := context.WithValue(c.Request.Context(), "request_id", rid)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
