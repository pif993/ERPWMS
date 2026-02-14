package http

import (
	"net/http"

	"erpwms/backend-go/internal/modules/admin/service"
	"github.com/gin-gonic/gin"
)

type AuthHandlers struct {
	Service      service.AuthService
	CookieSecure bool
}

type loginReq struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

func (h AuthHandlers) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	res, err := h.Service.Login(c, req.Email, req.Password, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	c.SetCookie("refresh_token", res.RefreshToken, 7*24*3600, "/", "", h.CookieSecure, true)
	if c.ContentType() == "application/x-www-form-urlencoded" {
		c.Redirect(http.StatusFound, "/stock")
		return
	}
	c.JSON(200, gin.H{"access_token": res.AccessToken})
}

func (h AuthHandlers) Refresh(c *gin.Context) {
	refresh, _ := c.Cookie("refresh_token")
	if refresh == "" {
		c.JSON(401, gin.H{"error": "missing refresh"})
		return
	}
	res, err := h.Service.Refresh(c, refresh, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid refresh"})
		return
	}
	c.SetCookie("refresh_token", res.RefreshToken, 7*24*3600, "/", "", h.CookieSecure, true)
	if c.ContentType() == "application/x-www-form-urlencoded" {
		c.Redirect(http.StatusFound, "/stock")
		return
	}
	c.JSON(200, gin.H{"access_token": res.AccessToken})
}

func (h AuthHandlers) Logout(c *gin.Context) {
	refresh, _ := c.Cookie("refresh_token")
	if refresh != "" {
		_ = h.Service.Logout(c, refresh, c.GetHeader("User-Agent"), c.ClientIP())
	}
	c.SetCookie("refresh_token", "", -1, "/", "", h.CookieSecure, true)
	c.Status(http.StatusNoContent)
}
