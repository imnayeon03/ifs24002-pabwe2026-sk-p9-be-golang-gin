package post

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"app/internal/middleware"
	"app/internal/response"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func uid(c *gin.Context) uint { return middleware.CurrentUser(c).UserID }

func id(c *gin.Context) (uint, bool) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "ID tidak valid")
		return 0, false
	}
	return uint(n), true
}

func (h *Handler) reply(c *gin.Context, code int, msg string, data interface{}, err error) {
	switch {
	case err == nil:
		response.Success(c, code, msg, data)
	case errors.Is(err, ErrNotFound):
		response.Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidCategory):
		response.Fail(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, "Terjadi kesalahan pada server")
	}
}

func (h *Handler) Categories(c *gin.Context) {
	response.Success(c, http.StatusOK, "Daftar kategori", Categories)
}

func (h *Handler) Public(c *gin.Context) {
	posts, err := h.svc.Public(Filter{Query: c.Query("q"), Category: c.Query("category")})
	h.reply(c, http.StatusOK, "Daftar artikel", posts, err)
}

func (h *Handler) PublicBySlug(c *gin.Context) {
	p, err := h.svc.PublicBySlug(c.Param("slug"))
	h.reply(c, http.StatusOK, "Detail artikel", p, err)
}

func (h *Handler) Mine(c *gin.Context) {
	posts, err := h.svc.Mine(uid(c))
	h.reply(c, http.StatusOK, "Artikel saya", posts, err)
}

func (h *Handler) Stats(c *gin.Context) {
	s, err := h.svc.Stats(uid(c))
	h.reply(c, http.StatusOK, "Statistik artikel", s, err)
}

func (h *Handler) Get(c *gin.Context) {
	n, ok := id(c)
	if !ok {
		return
	}
	p, err := h.svc.Get(uid(c), n)
	h.reply(c, http.StatusOK, "Detail artikel", p, err)
}

func (h *Handler) Create(c *gin.Context) {
	var req PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Judul artikel wajib diisi")
		return
	}
	p, err := h.svc.Create(uid(c), req)
	h.reply(c, http.StatusCreated, "Artikel dibuat", p, err)
}

func (h *Handler) Update(c *gin.Context) {
	n, ok := id(c)
	if !ok {
		return
	}
	var req PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Judul artikel wajib diisi")
		return
	}
	p, err := h.svc.Update(uid(c), n, req)
	h.reply(c, http.StatusOK, "Artikel diperbarui", p, err)
}

func (h *Handler) TogglePublish(c *gin.Context) {
	n, ok := id(c)
	if !ok {
		return
	}
	p, err := h.svc.TogglePublish(uid(c), n)
	h.reply(c, http.StatusOK, "Status publikasi diubah", p, err)
}

func (h *Handler) Delete(c *gin.Context) {
	n, ok := id(c)
	if !ok {
		return
	}
	h.reply(c, http.StatusOK, "Artikel dihapus", nil, h.svc.Delete(uid(c), n))
}

func (h *Handler) AdminUsers(c *gin.Context) {
	u, err := h.svc.AdminUsers()
	h.reply(c, http.StatusOK, "Daftar pengguna", u, err)
}

func (h *Handler) AdminPosts(c *gin.Context) {
	p, err := h.svc.AdminPosts()
	h.reply(c, http.StatusOK, "Semua artikel", p, err)
}

func (h *Handler) AdminUserPosts(c *gin.Context) {
	n, ok := id(c)
	if !ok {
		return
	}
	p, err := h.svc.AdminUserPosts(n)
	h.reply(c, http.StatusOK, "Artikel pengguna", p, err)
}
