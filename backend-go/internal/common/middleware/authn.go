package middleware

import (
	"strings"

	"erpwms/backend-go/internal/common/auth"
	"erpwms/backend-go/internal/db/sqlcgen"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func Authn(jwt auth.JWTManager, q *sqlcgen.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing bearer"})
			return
		}
		uid, err := jwt.Parse(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}
		var pgu pgtype.UUID
		if err := pgu.Scan(uid); err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid sub"})
			return
		}
		permsRows, err := q.ListPermissionsByUserID(c, pgu)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "permission load"})
			return
		}
		c.Set("user_id", uid)
		c.Set("permissions", permsRows)
		c.Next()
	}
}
