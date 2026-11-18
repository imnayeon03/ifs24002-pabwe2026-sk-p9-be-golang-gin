package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"app/internal/database"
	"app/internal/middleware"
)

func setup(t *testing.T) (*gin.Engine, *Service, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	db, err := database.ConnectMemory()
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))
	svc := NewService(NewRepository(db), "kunci")
	h := NewHandler(svc)
	r := gin.New()
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	a := r.Group("/", middleware.Auth("kunci"))
	a.POST("/logout", h.Logout)
	a.GET("/profile", h.Profile)
	a.PUT("/profile", h.UpdateProfile)
	return r, svc, db
}

func call(r *gin.Engine, method, path, token string, body interface{}) (*httptest.ResponseRecorder, map[string]interface{}) {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w, out
}

func TestAlurLengkap(t *testing.T) {
	r, _, _ := setup(t)
	reg := RegisterRequest{Username: "budi", Email: "b@x.id", Password: "rahasia"}
	w, _ := call(r, "POST", "/register", "", reg)
	assert.Equal(t, 201, w.Code)
	w, _ = call(r, "POST", "/register", "", reg)
	assert.Equal(t, http.StatusConflict, w.Code)
	w, _ = call(r, "POST", "/register", "", gin.H{"username": "x"})
	assert.Equal(t, 400, w.Code)

	w, _ = call(r, "POST", "/login", "", gin.H{})
	assert.Equal(t, 400, w.Code)
	w, _ = call(r, "POST", "/login", "", LoginRequest{Username: "budi", Password: "salah"})
	assert.Equal(t, 401, w.Code)
	w, _ = call(r, "POST", "/login", "", LoginRequest{Username: "tidak-ada", Password: "x"})
	assert.Equal(t, 401, w.Code)
	w, out := call(r, "POST", "/login", "", LoginRequest{Username: "budi", Password: "rahasia"})
	require.Equal(t, 200, w.Code)
	token := out["data"].(map[string]interface{})["token"].(string)

	w, out = call(r, "GET", "/profile", token, nil)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "budi", out["data"].(map[string]interface{})["username"])
	w, _ = call(r, "POST", "/logout", token, nil)
	assert.Equal(t, 200, w.Code)

	w, _ = call(r, "PUT", "/profile", token, UpdateProfileRequest{Username: "budi2", Email: "b2@x.id", Password: "baru123"})
	assert.Equal(t, 200, w.Code)
	w, _ = call(r, "POST", "/login", "", LoginRequest{Username: "budi2", Password: "baru123"})
	assert.Equal(t, 200, w.Code)
	w, _ = call(r, "PUT", "/profile", token, "bukan-objek")
	assert.Equal(t, 400, w.Code)
}

func TestProfilDuplikatDanTidakAda(t *testing.T) {
	r, svc, db := setup(t)
	_, _ = svc.Register(RegisterRequest{Username: "a", Email: "a@x.id", Password: "1234"})
	_, _ = svc.Register(RegisterRequest{Username: "b", Email: "b@x.id", Password: "1234"})
	_, err := svc.UpdateProfile(2, UpdateProfileRequest{Username: "a"})
	assert.ErrorIs(t, err, ErrDuplicate)
	_, err = svc.UpdateProfile(99, UpdateProfileRequest{})
	assert.ErrorIs(t, err, ErrNotFound)

	_, out := call(r, "POST", "/login", "", LoginRequest{Username: "a", Password: "1234"})
	token := out["data"].(map[string]interface{})["token"].(string)
	require.NoError(t, db.Delete(&database.User{}, 1).Error)
	w, _ := call(r, "GET", "/profile", token, nil)
	assert.Equal(t, 404, w.Code)
	w, _ = call(r, "PUT", "/profile", token, UpdateProfileRequest{})
	assert.Equal(t, 404, w.Code)
}

func TestGalatServer(t *testing.T) {
	r, svc, db := setup(t)
	_, _ = svc.Register(RegisterRequest{Username: "a", Email: "a@x.id", Password: "1234"})
	_, out := call(r, "POST", "/login", "", LoginRequest{Username: "a", Password: "1234"})
	token := out["data"].(map[string]interface{})["token"].(string)

	_, err := svc.Register(RegisterRequest{Username: "c", Email: "c@x.id", Password: string(make([]byte, 80))})
	assert.Error(t, err) // bcrypt menolak >72 byte
	_, err = svc.UpdateProfile(1, UpdateProfileRequest{Password: string(make([]byte, 80))})
	assert.Error(t, err)

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
	w, _ := call(r, "POST", "/register", "", RegisterRequest{Username: "z", Email: "z@x.id", Password: "1234"})
	assert.Equal(t, 500, w.Code)
	_, err = svc.UpdateProfile(1, UpdateProfileRequest{Username: "q"})
	assert.Error(t, err)
	_ = token
}
