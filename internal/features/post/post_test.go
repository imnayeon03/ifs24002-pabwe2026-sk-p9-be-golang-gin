package post

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"app/internal/database"
	"app/internal/middleware"
)

func setup(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	db, err := database.ConnectMemory()
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))
	for _, n := range []string{"u1", "u2"} {
		require.NoError(t, db.Create(&database.User{Username: n, Email: n + "@x.id", Password: "x"}).Error)
	}
	h := NewHandler(NewService(NewRepository(db)))
	r := gin.New()
	r.GET("/categories", h.Categories)
	r.GET("/public", h.Public)
	r.GET("/public/:slug", h.PublicBySlug)
	a := r.Group("/", func(c *gin.Context) {
		c.Set(middleware.ContextUser, middleware.Claims{UserID: 1, Username: "u1", Role: "admin"})
	})
	a.GET("/posts", h.Mine)
	a.POST("/posts", h.Create)
	a.GET("/posts/stats", h.Stats)
	a.GET("/posts/:id", h.Get)
	a.PUT("/posts/:id", h.Update)
	a.PATCH("/posts/:id/toggle-publish", h.TogglePublish)
	a.DELETE("/posts/:id", h.Delete)
	a.GET("/admin/users", h.AdminUsers)
	a.GET("/admin/posts", h.AdminPosts)
	a.GET("/admin/users/:id/posts", h.AdminUserPosts)
	return r, db
}

func call(r *gin.Engine, method, path string, body interface{}) (int, map[string]interface{}) {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var out map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func TestSlugifyDanReadingTime(t *testing.T) {
	assert.Equal(t, "belajar-go-gin-2026", Slugify("  Belajar Go & Gin 2026! "))
	assert.Equal(t, "artikel", Slugify("!!!"))
	assert.Equal(t, 1, ReadingTime(""))
	assert.Equal(t, 1, ReadingTime(strings.Repeat("kata ", 200)))
	assert.Equal(t, 2, ReadingTime(strings.Repeat("kata ", 201)))
}

func TestCRUDSlugDanPublikasi(t *testing.T) {
	r, _ := setup(t)
	code, out := call(r, "POST", "/posts", PostRequest{Title: "Halo Dunia", Category: "Teknologi", Content: "isi"})
	assert.Equal(t, 201, code)
	assert.Equal(t, "halo-dunia", out["data"].(map[string]interface{})["slug"])
	_, out = call(r, "POST", "/posts", PostRequest{Title: "Halo Dunia"})
	assert.Equal(t, "halo-dunia-2", out["data"].(map[string]interface{})["slug"])
	assert.Equal(t, "Umum", out["data"].(map[string]interface{})["category"])
	code, _ = call(r, "POST", "/posts", gin.H{})
	assert.Equal(t, 400, code)
	code, _ = call(r, "POST", "/posts", PostRequest{Title: "X", Category: "Ngawur"})
	assert.Equal(t, 400, code)

	code, _ = call(r, "GET", "/public/halo-dunia", nil)
	assert.Equal(t, 404, code) // masih draf
	code, out = call(r, "PATCH", "/posts/1/toggle-publish", nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, true, out["data"].(map[string]interface{})["is_published"])
	code, out = call(r, "GET", "/public/halo-dunia", nil)
	assert.Equal(t, 200, code)
	assert.EqualValues(t, 1, out["data"].(map[string]interface{})["reading_time"])

	code, _ = call(r, "PUT", "/posts/1", PostRequest{Title: "Judul Baru", Category: "Opini", Content: "x"})
	assert.Equal(t, 200, code)
	code, _ = call(r, "PUT", "/posts/abc", PostRequest{Title: "x"})
	assert.Equal(t, 400, code)
	code, _ = call(r, "PUT", "/posts/1", gin.H{})
	assert.Equal(t, 400, code)
	code, _ = call(r, "PUT", "/posts/99", PostRequest{Title: "x"})
	assert.Equal(t, 404, code)
	code, _ = call(r, "PUT", "/posts/1", PostRequest{Title: "x", Category: "Ngawur"})
	assert.Equal(t, 400, code)

	code, _ = call(r, "GET", "/posts/1", nil)
	assert.Equal(t, 200, code)
	code, _ = call(r, "GET", "/posts/abc", nil)
	assert.Equal(t, 400, code)
	code, _ = call(r, "GET", "/posts/99", nil)
	assert.Equal(t, 404, code)
	_, out = call(r, "GET", "/posts", nil)
	assert.Len(t, out["data"], 2)

	_, out = call(r, "GET", "/posts/stats", nil)
	s := out["data"].(map[string]interface{})
	assert.EqualValues(t, 2, s["total"])
	assert.EqualValues(t, 1, s["total_terbit"])
	assert.EqualValues(t, 1, s["belum_selesai"])

	code, out = call(r, "PATCH", "/posts/1/toggle-publish", nil) // kembali ke draf
	assert.Nil(t, out["data"].(map[string]interface{})["published_at"])
	code, _ = call(r, "PATCH", "/posts/abc/toggle-publish", nil)
	assert.Equal(t, 400, code)
	code, _ = call(r, "PATCH", "/posts/99/toggle-publish", nil)
	assert.Equal(t, 404, code)

	code, _ = call(r, "DELETE", "/posts/1", nil)
	assert.Equal(t, 200, code)
	code, _ = call(r, "DELETE", "/posts/abc", nil)
	assert.Equal(t, 400, code)
	code, _ = call(r, "DELETE", "/posts/1", nil)
	assert.Equal(t, 404, code)
}

func TestPublikCariFilterKategoriAdmin(t *testing.T) {
	r, db := setup(t)
	for i, p := range []PostRequest{
		{Title: "Belajar Golang", Category: "Teknologi", Content: "gin gonic"},
		{Title: "Resep Nasi", Category: "Gaya Hidup", Excerpt: "enak sekali"},
	} {
		call(r, "POST", "/posts", p)
		call(r, "PATCH", "/posts/"+string(rune('1'+i))+"/toggle-publish", nil)
	}
	_, out := call(r, "GET", "/public", nil)
	assert.Len(t, out["data"], 2)
	_, out = call(r, "GET", "/public?q=GOLANG", nil)
	assert.Len(t, out["data"], 1)
	_, out = call(r, "GET", "/public?q=enak", nil)
	assert.Len(t, out["data"], 1)
	_, out = call(r, "GET", "/public?category=Gaya%20Hidup", nil)
	assert.Len(t, out["data"], 1)
	_, out = call(r, "GET", "/public?q=tidakada", nil)
	assert.Len(t, out["data"], 0)

	_, out = call(r, "GET", "/categories", nil)
	assert.Len(t, out["data"], 6)

	_, out = call(r, "GET", "/admin/users", nil)
	u := out["data"].([]interface{})[0].(map[string]interface{})
	assert.EqualValues(t, 2, u["total_posts"])
	assert.EqualValues(t, 2, u["total_published"])
	_, out = call(r, "GET", "/admin/posts", nil)
	assert.Len(t, out["data"], 2)
	_, out = call(r, "GET", "/admin/users/1/posts", nil)
	assert.Len(t, out["data"], 2)
	code, _ := call(r, "GET", "/admin/users/abc/posts", nil)
	assert.Equal(t, 400, code)
	_ = db
}

func TestGalatDatabase(t *testing.T) {
	r, db := setup(t)
	call(r, "POST", "/posts", PostRequest{Title: "A"})
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
	for _, p := range [][2]string{
		{"GET", "/public"}, {"GET", "/posts"}, {"GET", "/posts/stats"},
		{"GET", "/admin/users"}, {"GET", "/admin/posts"}, {"GET", "/admin/users/1/posts"},
		{"POST", "/posts"}, {"GET", "/public/a"},
	} {
		code, _ := call(r, p[0], p[1], PostRequest{Title: "x"})
		assert.Contains(t, []int{404, 500}, code, p[1])
	}
}
