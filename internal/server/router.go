package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"app/docs"
	"app/internal/config"
	"app/internal/features/ai"
	"app/internal/features/auth"
	"app/internal/features/post"
	"app/internal/middleware"
	"app/internal/response"
)

// NewRouter merakit seluruh rute. chat boleh diganti (mock) saat pengujian.
func NewRouter(cfg config.Config, db *gorm.DB, chat ai.Chatter, search ai.Searcher) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS(), middleware.RateLimiter(600, time.Minute))

	authH := auth.NewHandler(auth.NewService(auth.NewRepository(db), cfg.AppKey))
	postH := post.NewHandler(post.NewService(post.NewRepository(db)))
	aiH := ai.NewHandler(ai.NewService(chat, search))

	r.GET("/api.json", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		b, err := docs.GetSwaggerJSON(c.Request.Host, scheme)
		if err != nil {
			response.Error(c, "Gagal membaca spesifikasi")
			return
		}
		c.Data(http.StatusOK, "application/json", b)
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/api.json")))

	api := r.Group("/api")
	api.GET("/public/categories", postH.Categories)
	api.GET("/public/posts", postH.Public)
	api.GET("/public/posts/:slug", postH.PublicBySlug)
	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)

	priv := api.Group("", middleware.Auth(cfg.AppKey))
	priv.POST("/auth/logout", authH.Logout)
	priv.GET("/auth/profile", authH.Profile)
	priv.PUT("/auth/profile", authH.UpdateProfile)
	priv.GET("/posts", postH.Mine)
	priv.POST("/posts", postH.Create)
	priv.GET("/posts/stats", postH.Stats)
	priv.POST("/posts/ai-generate", aiH.Generate)
	priv.GET("/posts/:id", postH.Get)
	priv.PUT("/posts/:id", postH.Update)
	priv.PATCH("/posts/:id/toggle-publish", postH.TogglePublish)
	priv.DELETE("/posts/:id", postH.Delete)

	adm := priv.Group("/admin", middleware.RequireRole("admin"))
	adm.GET("/users", postH.AdminUsers)
	adm.GET("/posts", postH.AdminPosts)
	adm.GET("/users/:id/posts", postH.AdminUserPosts)

	r.NoRoute(func(c *gin.Context) { response.Fail(c, http.StatusNotFound, "Endpoint not found") })
	return r
}
