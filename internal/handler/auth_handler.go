package handler

import (
	"github.com/gin-gonic/gin"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/dto"
	"warehouse-management/internal/service"
	"warehouse-management/pkg/response"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := bindJSON(c, &req); err != nil {
		response.Error(c, err)
		return
	}

	result, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, dto.LoginResponse{
		AccessToken: result.Token,
		TokenType:   "Bearer",
		ExpiresAt:   result.ExpiresAt,
		User:        dto.NewMeResponse(result.User),
	})
}

// GET /api/v1/auth/me (wajib login, lewat middleware.Authenticate)
func (h *AuthHandler) Me(c *gin.Context) {
	response.OK(c, dto.NewMeResponse(auth.CurrentUser(c)))
}
