package http

import (
	"net/http"
	"os"

	"erpwms/backend-go/internal/modules/autotest/service"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Enabled bool
	Token   string
	Router  http.Handler
}

func (h Handlers) Page(c *gin.Context) {
	if !h.Enabled {
		c.Status(http.StatusNotFound)
		return
	}
	c.HTML(http.StatusOK, "pages/autotest.html", nil)
}

func (h Handlers) Run(c *gin.Context) {
	if !h.Enabled {
		c.JSON(http.StatusNotFound, gin.H{"error": "disabled"})
		return
	}
	if h.Token == "" || c.GetHeader("X-Autotest-Token") != h.Token {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminEmail == "" || adminPassword == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing ADMIN_EMAIL/ADMIN_PASSWORD"})
		return
	}

	res := service.Run(h.Router, adminEmail, adminPassword)
	c.JSON(http.StatusOK, res)
}
