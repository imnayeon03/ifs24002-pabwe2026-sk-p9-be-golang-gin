package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"app/internal/middleware"
	"app/internal/response"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrDuplicate):
		response.Fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrCredentials):
		response.Fail(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrNotFound):
		response.Fail(c, http.StatusNotFound, err.Error())
	default:
		response.Error(c, "Terjadi kesalahan pada server")
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Data registrasi tidak valid")
		return
	}
	u, err := h.svc.Register(req)
	if err != nil {
		h.fail(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Registrasi berhasil", u)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Username dan password wajib diisi")
		return
	}
	token, u, err := h.svc.Login(req)
	if err != nil {
		h.fail(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Login berhasil", gin.H{"token": token, "user": u})
}

func (h *Handler) Logout(c *gin.Context) {
	response.Success(c, http.StatusOK, "Logout berhasil", nil)
}

func (h *Handler) Profile(c *gin.Context) {
	u, err := h.svc.Profile(middleware.CurrentUser(c).UserID)
	if err != nil {
		h.fail(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Profil pengguna", u)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Data profil tidak valid")
		return
	}
	u, err := h.svc.UpdateProfile(middleware.CurrentUser(c).UserID, req)
	if err != nil {
		h.fail(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Profil diperbarui", u)
}
