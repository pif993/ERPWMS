package middleware

import (
	"strings"

	"erpwms/backend-go/internal/common/auth"
	sqlcgen "erpwms/backend-go/internal/db/sqlcgen"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func Authn(jwtMgr auth.JWTManager, q *sqlcgen.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		userID, err := jwtMgr.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}

		parsed, err := uuid.Parse(userID)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		var userUUID pgtype.UUID
		if err := userUUID.Scan(parsed.String()); err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}

		permissions, err := q.ListPermissionsByUserID(c, userUUID)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}

		c.Set("user_id", userID)
		c.Set("permissions", permissions)
		c.Next()
	}
}
