// Package dto: bentuk request & response API, dipisah dari model tabel.
package dto

import (
	"time"

	"github.com/google/uuid"

	"warehouse-management/internal/auth"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string     `json:"access_token"`
	TokenType   string     `json:"token_type"`
	ExpiresAt   time.Time  `json:"expires_at"`
	User        MeResponse `json:"user"`
}

type MeResponse struct {
	ID            uuid.UUID   `json:"id"`
	Name          string      `json:"name"`
	Email         string      `json:"email"`
	Roles         []string    `json:"roles"`
	Permissions   []string    `json:"permissions"`
	AllWarehouses bool        `json:"all_warehouses"`
	WarehouseIDs  []uuid.UUID `json:"warehouse_ids"`
}

func NewMeResponse(u *auth.User) MeResponse {
	ids := u.WarehouseIDs
	if ids == nil {
		ids = []uuid.UUID{} // render [] bukan null
	}
	return MeResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Roles:         u.Roles,
		Permissions:   u.Permissions(),
		AllWarehouses: u.AllWarehouses,
		WarehouseIDs:  ids,
	}
}
