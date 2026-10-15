package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/repository"
	"warehouse-management/pkg/apperror"
)

type AuthService struct {
	users  *repository.UserRepository
	tokens *auth.TokenManager
}

func NewAuthService(users *repository.UserRepository, tokens *auth.TokenManager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	User      *auth.User
}

// Pesan sengaja sama untuk "email tidak ada" dan "password salah".
var errInvalidCredential = apperror.Unauthorized("email atau password salah")

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errInvalidCredential
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errInvalidCredential
	}
	if !user.IsActive {
		return nil, apperror.Forbidden("akun tidak aktif")
	}

	token, expiresAt, err := s.tokens.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	authUser, err := s.LoadUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, ExpiresAt: expiresAt, User: authUser}, nil
}

// LoadUser membentuk auth.User lengkap (role, permission, gudang).
func (s *AuthService) LoadUser(ctx context.Context, userID uuid.UUID) (*auth.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.Unauthorized("user tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, apperror.Forbidden("akun tidak aktif")
	}

	roles, err := s.users.GetRoleNames(ctx, userID)
	if err != nil {
		return nil, err
	}
	perms, err := s.users.GetPermissionNames(ctx, userID)
	if err != nil {
		return nil, err
	}
	warehouseIDs, err := s.users.GetWarehouseIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	return auth.NewUser(user.ID, user.Name, user.Email, roles, perms, warehouseIDs), nil
}
