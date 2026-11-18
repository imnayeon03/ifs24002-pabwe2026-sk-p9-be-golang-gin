package ai

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"app/internal/response"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Topik wajib diisi")
		return
	}
	d, err := h.svc.Generate(req)
	if err != nil {
		response.Error(c, "Gagal menghasilkan artikel: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Draf artikel berhasil dibuat", d)
}
