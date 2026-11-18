package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func token(t *testing.T, secret, role string) string {
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{"uid": 7, "username": "u", "role": role}).SignedString([]byte(secret))
	assert.NoError(t, err)
	return s
}

func do(r *gin.Engine, method, header string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, "/x", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })
	r.OPTIONS("/x", func(c *gin.Context) {})
	assert.Equal(t, 204, do(r, "OPTIONS", "").Code)
	w := do(r, "GET", "")
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestAuthDanRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", Auth("k"), RequireRole("admin"), func(c *gin.Context) {
		c.JSON(200, CurrentUser(c))
	})
	assert.Equal(t, 401, do(r, "GET", "").Code)
	assert.Equal(t, 401, do(r, "GET", "Basic abc").Code)
	assert.Equal(t, 401, do(r, "GET", "Bearer rusak").Code)
	assert.Equal(t, 401, do(r, "GET", "Bearer "+token(t, "lain", "admin")).Code)
	assert.Equal(t, 403, do(r, "GET", "Bearer "+token(t, "k", "member")).Code)
	w := do(r, "GET", "Bearer "+token(t, "k", "admin"))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"Username":"u"`)
}

func TestCurrentUserKosong(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Equal(t, Claims{}, CurrentUser(c))
}

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimiter(2, 50*time.Millisecond))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	assert.Equal(t, 200, do(r, "GET", "").Code)
	assert.Equal(t, 200, do(r, "GET", "").Code)
	assert.Equal(t, 429, do(r, "GET", "").Code)
	time.Sleep(70 * time.Millisecond)
	assert.Equal(t, 200, do(r, "GET", "").Code)
}
