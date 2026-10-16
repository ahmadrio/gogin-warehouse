package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/dto"
	"warehouse-management/internal/model"
	"warehouse-management/internal/repository"
	"warehouse-management/pkg/apperror"
	"warehouse-management/pkg/pagination"
)

type WarehouseService struct {
	db         *gorm.DB
	warehouses *repository.WarehouseRepository
	audits     *repository.AuditRepository
}

func NewWarehouseService(db *gorm.DB, warehouses *repository.WarehouseRepository, audits *repository.AuditRepository) *WarehouseService {
	return &WarehouseService{db: db, warehouses: warehouses, audits: audits}
}

// List: user.WarehouseScope() diteruskan ke repository (lapis kedua hak akses).
func (s *WarehouseService) List(ctx context.Context, user *auth.User, p pagination.Params) ([]model.Warehouse, int64, error) {
	return s.warehouses.List(ctx, user.WarehouseScope(), p)
}

func (s *WarehouseService) Get(ctx context.Context, user *auth.User, id uuid.UUID) (*model.Warehouse, error) {
	w, err := s.warehouses.FindByID(ctx, user.WarehouseScope(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("warehouse")
	}
	return w, err
}

func (s *WarehouseService) ListLocations(ctx context.Context, user *auth.User, warehouseID uuid.UUID) ([]model.WarehouseLocation, error) {
	// Cek akses ke gudangnya dulu, baru ambil lokasinya.
	if _, err := s.Get(ctx, user, warehouseID); err != nil {
		return nil, err
	}
	return s.warehouses.ListLocations(ctx, warehouseID)
}

func (s *WarehouseService) Create(ctx context.Context, user *auth.User, req dto.CreateWarehouseRequest) (*model.Warehouse, error) {
	exists, err := s.warehouses.ExistsByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.Conflict("kode warehouse sudah dipakai")
	}

	w := &model.Warehouse{Code: req.Code, Name: req.Name, Address: req.Address, IsActive: true}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.warehouses.WithTx(tx).Create(ctx, w); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionCreate, model.EntityWarehouse,
			w.ID, fmt.Sprintf("membuat warehouse %s", w.Code)))
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, apperror.Conflict("kode warehouse sudah dipakai")
	}
	if err != nil {
		return nil, err
	}
	return w, nil
}
