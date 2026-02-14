package middleware

import (
	"net/http"
	"strings"

	"erpwms/backend-go/internal/common/auth"
	sqlc "erpwms/backend-go/internal/db/sqlcgen"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func Authn(jwtMgr auth.JWTManager, q *sqlc.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		userIDStr, err := jwtMgr.Parse(token)
		if err != nil || userIDStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		var uid pgtype.UUID
		if err := uid.Scan(userIDStr); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		perms, err := q.ListPermissionsByUserID(c, uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("user_id", userIDStr)
		c.Set("permissions", perms)
		c.Next()
	}
}
