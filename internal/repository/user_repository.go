package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"warehouse-management/internal/model"
)

// ErrNotFound dipakai semua repository saat data tidak ada.
var ErrNotFound = errors.New("record not found")

func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("lower(email) = lower(?)", email).First(&user).Error
	return &user, notFound(err)
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	return &user, notFound(err)
}

// GetRoleNames: nama role milik user, mis. ["staff"].
func (r *UserRepository) GetRoleNames(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).
		Table("user_roles ur").
		Joins("JOIN roles r ON r.id = ur.role_id AND r.deleted_at IS NULL").
		Where("ur.user_id = ? AND ur.deleted_at IS NULL", userID).
		Pluck("r.name", &names).Error
	return names, err
}

// GetPermissionNames: semua permission dari semua role user, format "resource:operation".
func (r *UserRepository) GetPermissionNames(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).
		Table("user_roles ur").
		Joins("JOIN roles r ON r.id = ur.role_id AND r.deleted_at IS NULL").
		Joins("JOIN role_permissions rp ON rp.role_id = ur.role_id AND rp.deleted_at IS NULL").
		Joins("JOIN permissions p ON p.id = rp.permission_id AND p.deleted_at IS NULL").
		Joins("JOIN resources res ON res.id = p.resource_id AND res.deleted_at IS NULL").
		Joins("JOIN operations op ON op.id = p.operation_id AND op.deleted_at IS NULL").
		Where("ur.user_id = ? AND ur.deleted_at IS NULL", userID).
		Distinct().
		Pluck("res.name || ':' || op.name", &names).Error
	return names, err
}

// GetWarehouseIDs: gudang yang boleh diakses user.
func (r *UserRepository) GetWarehouseIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).
		Table("user_warehouses uw").
		Joins("JOIN warehouses w ON w.id = uw.warehouse_id AND w.deleted_at IS NULL").
		Where("uw.user_id = ? AND uw.deleted_at IS NULL", userID).
		Pluck("uw.warehouse_id", &ids).Error
	return ids, err
}
