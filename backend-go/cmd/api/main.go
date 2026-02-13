package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"erpwms/backend-go/internal/common/auth"
	"erpwms/backend-go/internal/common/config"
	"erpwms/backend-go/internal/common/crypto"
	"erpwms/backend-go/internal/common/middleware"
	sqlc "erpwms/backend-go/internal/db/sqlcgen"
	adminhttp "erpwms/backend-go/internal/modules/admin/http"
	adminsvc "erpwms/backend-go/internal/modules/admin/service"
	stockhttp "erpwms/backend-go/internal/modules/wms_stock/http"
	stocksvc "erpwms/backend-go/internal/modules/wms_stock/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	redis "github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := pgxpool.New(context.Background(), cfg.DBURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	q := sqlc.New(db)

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	nc, _ := nats.Connect(cfg.NATSURL)
	jwtMgr := auth.JWTManager{Issuer: cfg.JWTIssuer, Audience: cfg.JWTAudience, Current: []byte(cfg.JWTCurrent), Previous: []byte(cfg.JWTPrevious)}
	authSvc := adminsvc.AuthService{Queries: q, JWT: jwtMgr, SearchPepper: cfg.SearchPepper, AuditPepper: cfg.AuditPepper, Argon: crypto.DefaultArgon2Params()}
	stockSvc := stocksvc.StockService{DB: db, Queries: q}

	r := gin.New()
	r.LoadHTMLGlob("web/templates/**/*.html")
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.SecurityHeaders(), middleware.CORS(cfg.CorsOrigins), middleware.RateLimit(cfg.RateLimitAPI))

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
		logger.Error("health degraded", "db", dberr, "redis", rerr, "nats_ok", nok)
		c.JSON(503, gin.H{"status": "degraded"})
	})

	ah := adminhttp.AuthHandlers{Service: authSvc, CookieSecure: cfg.CookieSecure}
	r.GET("/login", func(c *gin.Context) { c.HTML(200, "pages/login.html", nil) })
	r.POST("/login", ah.Login)
	r.GET("/stock", func(c *gin.Context) { c.HTML(200, "pages/stock.html", nil) })

	api := r.Group("/api")
	api.POST("/auth/login", ah.Login)
	api.POST("/auth/refresh", ah.Refresh)
	api.POST("/auth/logout", ah.Logout)

	authed := api.Group("/")
	authed.Use(middleware.Authn(jwtMgr, q))
	sh := stockhttp.StockHandlers{Queries: q, Service: stockSvc}
	authed.GET("stock/balances", middleware.RequirePermission("wms.stock.read"), sh.ListBalances)
	authed.POST("stock/moves", middleware.RequirePermission("wms.stock.move"), sh.Move)

	if err := r.Run(cfg.HTTPAddr); err != nil {
		panic(err)
	}
}
