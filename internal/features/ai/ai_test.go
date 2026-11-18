package ai

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func fakeServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

const okBody = `{"choices":[{"message":{"content":"{\"title\":\"T\"}"}}]}`

func TestClientChat(t *testing.T) {
	s := fakeServer(200, okBody)
	defer s.Close()
	got, err := NewClient(s.URL+"/", "k", "m").Chat("s", "u")
	assert.NoError(t, err)
	assert.Contains(t, got, "title")
	for _, c := range []struct {
		status int
		body   string
	}{{500, "x"}, {200, "bukan json"}, {200, `{"choices":[]}`}} {
		bad := fakeServer(c.status, c.body)
		_, err = NewClient(bad.URL, "k", "m").Chat("s", "u")
		assert.Error(t, err)
		bad.Close()
	}
	_, err = NewClient("http://127.0.0.1:1", "k", "m").Chat("s", "u")
	assert.Error(t, err)
	_, err = NewClient("http://bad host", "k", "m").Chat("s", "u")
	assert.Error(t, err)
}

func TestExtractJSON(t *testing.T) {
	assert.Equal(t, `{"a":1}`, ExtractJSON("```json\n{\"a\":1}\n```"))
	assert.Equal(t, "polos", ExtractJSON(" polos "))
}

func TestSearch(t *testing.T) {
	for _, body := range []string{
		`{"results":[{"title":" A ","url":"http://a","snippet":"s"}]}`,
		`{"data":[{"title":"A","link":"http://a","content":"s"}]}`,
		`{"data":{"results":[{"title":"A","url":"http://a"}]}}`,
	} {
		s := fakeServer(200, body)
		got, err := NewSearchClient(s.URL).Search("go lang")
		assert.NoError(t, err)
		if assert.Len(t, got, 1) {
			assert.Equal(t, "http://a", got[0].URL)
		}
		s.Close()
	}
	s := fakeServer(200, `{"results":[{"title":"tanpa url"}]}`)
	got, _ := NewSearchClient(s.URL).Search("x")
	assert.Empty(t, got)
	s.Close()
	s = fakeServer(200, `{"data":"bukan-array"}`)
	got, _ = NewSearchClient(s.URL).Search("x")
	assert.Empty(t, got)
	s.Close()
	s = fakeServer(500, "x")
	_, err := NewSearchClient(s.URL).Search("x")
	assert.Error(t, err)
	s.Close()
	s = fakeServer(200, "bukan json")
	_, err = NewSearchClient(s.URL).Search("x")
	assert.Error(t, err)
	s.Close()
	_, err = NewSearchClient("http://127.0.0.1:1").Search("x")
	assert.Error(t, err)
}

type stubChat struct {
	out string
	err error
}

func (s stubChat) Chat(string, string) (string, error) { return s.out, s.err }

type stubSearch struct {
	src []Source
	err error
}

func (s stubSearch) Search(string) ([]Source, error) { return s.src, s.err }

const goodDraft = `{"title":"Judul","category":"","excerpt":"e","content":"# isi"}`

func TestGenerate(t *testing.T) {
	src := []Source{{Title: "S", URL: "http://s", Snippet: "x"}}
	d, err := NewService(stubChat{out: goodDraft}, stubSearch{src: src}).Generate(GenerateRequest{Topic: "t", Category: "Opini", IncludeWebSearch: true})
	assert.NoError(t, err)
	assert.Equal(t, "Opini", d.Category)
	assert.Equal(t, src, d.Sources)

	d, err = NewService(stubChat{out: goodDraft}, stubSearch{err: errors.New("mati")}).Generate(GenerateRequest{Topic: "t", IncludeWebSearch: true})
	assert.NoError(t, err)
	assert.Empty(t, d.Sources)

	d, err = NewService(stubChat{out: goodDraft}, stubSearch{src: src}).Generate(GenerateRequest{Topic: "t"})
	assert.NoError(t, err)
	assert.Empty(t, d.Sources)

	_, err = NewService(stubChat{err: errors.New("x")}, stubSearch{}).Generate(GenerateRequest{Topic: "t"})
	assert.Error(t, err)
	_, err = NewService(stubChat{out: "ngawur"}, stubSearch{}).Generate(GenerateRequest{Topic: "t"})
	assert.ErrorIs(t, err, ErrBadAnswer)
	_, err = NewService(stubChat{out: `{"title":"x"}`}, stubSearch{}).Generate(GenerateRequest{Topic: "t"})
	assert.ErrorIs(t, err, ErrBadAnswer)
}

func TestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/g", NewHandler(NewService(stubChat{out: goodDraft}, stubSearch{})).Generate)
	r.POST("/e", NewHandler(NewService(stubChat{err: errors.New("x")}, stubSearch{})).Generate)
	post := func(p, b string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", p, bytes.NewBufferString(b))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		return w.Code
	}
	assert.Equal(t, 200, post("/g", `{"topic":"x"}`))
	assert.Equal(t, 400, post("/g", `{}`))
	assert.Equal(t, 500, post("/e", `{"topic":"x"}`))
}
