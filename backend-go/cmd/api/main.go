package main

import (
	"context"
	"time"

	"erpwms/backend-go/internal/common/config"
	"erpwms/backend-go/internal/common/middleware"
	"erpwms/backend-go/internal/common/rbac"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	redis "github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	db, err := pgxpool.New(context.Background(), cfg.DBURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379"})
	nc, _ := nats.Connect(cfg.NATSURL)

	r := gin.New()
	r.LoadHTMLGlob("web/templates/**/*.html")
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) { c.Writer.Header().Set("X-Content-Type-Options", "nosniff"); c.Next() })
	r.Use(func(c *gin.Context) { c.Set("permissions", []string{"wms.stock.read", "wms.stock.move"}); c.Next() })

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 2*time.Second)
		defer cancel()
		dberr := db.Ping(ctx)
		rerr := rdb.Ping(ctx).Err()
		nok := nc != nil && nc.Status() == nats.CONNECTED
		if dberr == nil && rerr == nil && nok {
			c.JSON(200, gin.H{"status": "ok"})
			return
		}
		c.JSON(503, gin.H{"status": "degraded"})
	})
	r.GET("/login", func(c *gin.Context) { c.HTML(200, "pages/login.html", gin.H{}) })
	r.POST("/login", func(c *gin.Context) { c.Redirect(302, "/") })
	r.GET("/", func(c *gin.Context) { c.HTML(200, "pages/dashboard.html", gin.H{"Title": "Dashboard"}) })
	r.GET("/stock", func(c *gin.Context) { c.HTML(200, "pages/stock.html", gin.H{"Title": "Stock"}) })

	r.POST("/api/auth/login", func(c *gin.Context) {
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "user-1", "exp": time.Now().Add(15 * time.Minute).Unix(), "iss": cfg.JWTIssuer, "aud": cfg.JWTAudience})
		s, _ := tok.SignedString([]byte(cfg.JWTKeyCurrent))
		c.JSON(200, gin.H{"access_token": s, "refresh_token": "placeholder"})
	})
	r.POST("/api/auth/refresh", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/api/auth/logout", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/api/stock/balances", func(c *gin.Context) { c.JSON(200, gin.H{"items": []any{}}) })
	r.POST("/api/stock/moves", rbac.RequirePermission("wms.stock.move"), func(c *gin.Context) {
		k := c.GetHeader("Idempotency-Key")
		if k == "" {
			c.JSON(400, gin.H{"error": "missing Idempotency-Key"})
			return
		}
		h := middleware.HashPayload(k)
		_, _ = db.Exec(c, "INSERT INTO audit_log(actor_type,action,resource,status,request_id) VALUES('user','stock.move','stock_ledger','ok',$1)", h)
		_, _ = db.Exec(c, "INSERT INTO outbox_events(subject,payload) VALUES('stock.moved', $1::jsonb)", `{"idempotency":"`+h+`"}`)
		c.JSON(200, gin.H{"status": "moved", "idempotency_hash": h})
	})
	r.POST("/api/orders", func(c *gin.Context) { c.JSON(201, gin.H{"status": "created"}) })
	r.POST("/api/orders/:id/allocate", func(c *gin.Context) { c.JSON(200, gin.H{"status": "allocated"}) })

	r.Run(cfg.HTTPAddr)
}
