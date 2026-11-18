package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"app/internal/response"
)

const ContextUser = "user"

type Claims struct {
	UserID   uint
	Username string
	Role     string
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Auth memverifikasi Bearer JWT (HS256) dan menaruh Claims di context.
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if raw == "" || raw == c.GetHeader("Authorization") {
			response.AbortFail(c, http.StatusUnauthorized, "Token tidak ditemukan")
			return
		}
		tok, err := jwt.Parse(raw, func(*jwt.Token) (interface{}, error) { return []byte(secret), nil },
			jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !tok.Valid {
			response.AbortFail(c, http.StatusUnauthorized, "Token tidak valid")
			return
		}
		m := tok.Claims.(jwt.MapClaims)
		id, _ := m["uid"].(float64)
		name, _ := m["username"].(string)
		role, _ := m["role"].(string)
		c.Set(ContextUser, Claims{UserID: uint(id), Username: name, Role: role})
		c.Next()
	}
}

// CurrentUser mengambil Claims hasil Auth.
func CurrentUser(c *gin.Context) Claims {
	v, _ := c.Get(ContextUser)
	cl, _ := v.(Claims)
	return cl
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := CurrentUser(c).Role
		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}
		response.AbortFail(c, http.StatusForbidden, "Akses ditolak")
	}
}

// RateLimiter: maksimum `limit` request per IP per `window`.
func RateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	hits := map[string][]time.Time{}
	return func(c *gin.Context) {
		now := time.Now()
		mu.Lock()
		recent := hits[c.ClientIP()][:0]
		for _, t := range hits[c.ClientIP()] {
			if now.Sub(t) < window {
				recent = append(recent, t)
			}
		}
		over := len(recent) >= limit
		if !over {
			recent = append(recent, now)
		}
		hits[c.ClientIP()] = recent
		mu.Unlock()
		if over {
			response.AbortFail(c, http.StatusTooManyRequests, "Terlalu banyak permintaan")
			return
		}
		c.Next()
	}
}
