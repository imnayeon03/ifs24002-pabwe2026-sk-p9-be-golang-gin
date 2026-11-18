package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func run(f func(c *gin.Context)) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	f(c)
	return w
}

func TestSemuaHelper(t *testing.T) {
	w := run(func(c *gin.Context) { Success(c, 200, "ok", gin.H{"a": 1}) })
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"success"`)
	w = run(func(c *gin.Context) { Fail(c, 400, "x") })
	assert.Contains(t, w.Body.String(), `"status":"fail"`)
	assert.NotContains(t, w.Body.String(), `"data"`)
	w = run(func(c *gin.Context) { Error(c, "x") })
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	w = run(func(c *gin.Context) { AbortFail(c, 401, "x") })
	assert.Equal(t, 401, w.Code)
	w = run(func(c *gin.Context) { AbortError(c, "x") })
	assert.Contains(t, w.Body.String(), `"status":"error"`)
}
