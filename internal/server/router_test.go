package server

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"app/internal/config"
	"app/internal/database"
	"app/internal/features/ai"
)

type fakeChat struct{}

func (fakeChat) Chat(string, string) (string, error) {
	return `{"title":"Judul AI","category":"Teknologi","excerpt":"e","content":"isi"}`, nil
}

type fakeSearch struct{}

func (fakeSearch) Search(string) ([]ai.Source, error) {
	return []ai.Source{{Title: "S", URL: "http://s"}}, nil
}

func newApp(t *testing.T) *gin.Engine {
	db, err := database.ConnectMemory()
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))
	cfg := config.Config{AppEnv: "production", AppKey: "kunci", AdminUsername: "admin", AdminPassword: "admin"}
	require.NoError(t, database.Seed(db, cfg))
	return NewRouter(cfg, db, fakeChat{}, fakeSearch{})
}

func do(r *gin.Engine, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func login(t *testing.T, r *gin.Engine, u, p string) string {
	w := do(r, "POST", "/api/auth/login", "", `{"username":"`+u+`","password":"`+p+`"}`)
	require.Equal(t, 200, w.Code)
	var out struct{ Data struct{ Token string } }
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	return out.Data.Token
}

func TestRuteEndToEnd(t *testing.T) {
	r := newApp(t)
	assert.Equal(t, 404, do(r, "GET", "/tidak-ada", "", "").Code)
	assert.Equal(t, 401, do(r, "GET", "/api/posts", "", "").Code)
	assert.Equal(t, 200, do(r, "GET", "/api/public/categories", "", "").Code)

	member := login(t, r, "member", "member")
	admin := login(t, r, "admin", "admin")
	assert.Equal(t, 201, do(r, "POST", "/api/posts", member, `{"title":"Halo"}`).Code)
	assert.Equal(t, 200, do(r, "PATCH", "/api/posts/1/toggle-publish", member, "").Code)
	assert.Equal(t, 200, do(r, "GET", "/api/public/posts", "", "").Code)
	assert.Equal(t, 200, do(r, "GET", "/api/public/posts/halo", "", "").Code)
	assert.Equal(t, 200, do(r, "GET", "/api/posts/stats", member, "").Code)
	w := do(r, "POST", "/api/posts/ai-generate", member, `{"topic":"x","include_web_search":true}`)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "http://s")
	assert.Equal(t, 403, do(r, "GET", "/api/admin/users", member, "").Code)
	assert.Equal(t, 200, do(r, "GET", "/api/admin/users", admin, "").Code)
	assert.Equal(t, 200, do(r, "GET", "/api/admin/posts", admin, "").Code)
	assert.Equal(t, 200, do(r, "GET", "/api/admin/users/2/posts", admin, "").Code)
}

func TestSwagger(t *testing.T) {
	r := newApp(t)
	w := do(r, "GET", "/api.json", "", "")
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "Delcom Blog API")
	req := httptest.NewRequest("GET", "/api.json", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Contains(t, w.Body.String(), `"https"`)
	assert.Equal(t, 200, do(r, "GET", "/swagger/index.html", "", "").Code)
}
