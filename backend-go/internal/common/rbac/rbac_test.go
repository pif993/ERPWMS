package rbac

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("permissions", []string{"wms.stock.read"}); c.Next() })
	r.GET("/", RequirePermission("wms.stock.read"), func(c *gin.Context) { c.Status(200) })
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal("expected 200")
	}
}
