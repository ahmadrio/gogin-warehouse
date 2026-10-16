package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/service"
	"warehouse-management/pkg/apperror"
	"warehouse-management/pkg/response"
)

// Authenticate: wajib login.
// Alur: ambil "Authorization: Bearer <token>" -> validasi JWT -> muat user + hak akses dari DB
// -> simpan ke context -> lanjut ke handler berikutnya.
func Authenticate(tokens *auth.TokenManager, authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenString, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || tokenString == "" {
			response.Error(c, apperror.Unauthorized("token tidak ditemukan"))
			return
		}

		userID, err := tokens.Parse(tokenString)
		if err != nil {
			response.Error(c, apperror.Unauthorized("token tidak valid atau sudah kedaluwarsa"))
			return
		}

		user, err := authService.LoadUser(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		auth.SetUser(c, user)
		c.Next()
	}
}

// RequirePermission: cek permission fitur, contoh RequirePermission("inbounds:approve").
// Lolos di sini BELUM berarti boleh mengakses semua gudang;
// pembatasan gudang dilakukan di service/repository lewat WarehouseScope.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !auth.CurrentUser(c).Can(permission) {
			response.Error(c, apperror.Forbidden("anda tidak memiliki akses: "+permission))
			return
		}
		c.Next()
	}
}
