package http

import (
	"strconv"

	"erpwms/backend-go/internal/db/sqlcgen"
	"erpwms/backend-go/internal/modules/wms_stock/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StockHandlers struct {
	Queries *sqlcgen.Queries
	Service service.StockService
}

func (h StockHandlers) ListBalances(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)
	rows, err := h.Queries.ListStockBalances(c.Request.Context(), sqlcgen.ListStockBalancesParams{
		Column1: c.Query("q"),
		Column2: c.Query("warehouse"),
		Column3: c.Query("location"),
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "db"})
		return
	}
	c.JSON(200, gin.H{"items": rows})
}

func (h StockHandlers) Move(c *gin.Context) {
	key := c.GetHeader("Idempotency-Key")
	if key == "" {
		c.JSON(400, gin.H{"error": "Idempotency-Key required"})
		return
	}
	var req service.MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	uid, _ := uuid.Parse(c.GetString("user_id"))
	resp, err := h.Service.MoveStock(c.Request.Context(), req, uid, "/api/stock/moves", key)
	if err != nil {
		c.JSON(409, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, resp)
}
