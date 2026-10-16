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
)

type WarehouseLocationService struct {
	db         *gorm.DB
	locations  *repository.WarehouseLocationRepository
	warehouses *repository.WarehouseRepository
	audits     *repository.AuditRepository
}

func NewWarehouseLocationService(db *gorm.DB, locations *repository.WarehouseLocationRepository, warehouses *repository.WarehouseRepository, audits *repository.AuditRepository) *WarehouseLocationService {
	return &WarehouseLocationService{db: db, locations: locations, warehouses: warehouses, audits: audits}
}

func (s *WarehouseLocationService) Get(ctx context.Context, user *auth.User, warehouseID, id uuid.UUID) (*model.WarehouseLocation, error) {
	if err := s.checkWarehouse(ctx, user, warehouseID); err != nil {
		return nil, err
	}
	loc, err := s.locations.FindByID(ctx, warehouseID, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("lokasi")
	}
	return loc, err
}

func (s *WarehouseLocationService) Create(ctx context.Context, user *auth.User, warehouseID uuid.UUID, req dto.CreateWarehouseLocationRequest) (*model.WarehouseLocation, error) {
	if err := s.checkWarehouse(ctx, user, warehouseID); err != nil {
		return nil, err
	}
	if err := s.checkCode(ctx, warehouseID, req.Code, nil); err != nil {
		return nil, err
	}

	loc := &model.WarehouseLocation{
		WarehouseID: warehouseID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive == nil || *req.IsActive,
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.locations.WithTx(tx).Create(ctx, loc); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionCreate, model.EntityWarehouseLocation,
			loc.ID, fmt.Sprintf("membuat lokasi %s", loc.Code)))
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, apperror.Conflict("kode lokasi sudah dipakai di warehouse ini")
	}
	if err != nil {
		return nil, err
	}
	return loc, nil
}

func (s *WarehouseLocationService) Update(ctx context.Context, user *auth.User, warehouseID, id uuid.UUID, req dto.UpdateWarehouseLocationRequest) (*model.WarehouseLocation, error) {
	loc, err := s.Get(ctx, user, warehouseID, id)
	if err != nil {
		return nil, err
	}
	if err := s.checkCode(ctx, warehouseID, req.Code, &id); err != nil {
		return nil, err
	}

	loc.Code = req.Code
	loc.Name = req.Name
	loc.Description = req.Description
	loc.IsActive = *req.IsActive

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.locations.WithTx(tx).Update(ctx, loc); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionUpdate, model.EntityWarehouseLocation,
			loc.ID, fmt.Sprintf("mengubah lokasi %s", loc.Code)))
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, apperror.Conflict("kode lokasi sudah dipakai di warehouse ini")
	}
	if err != nil {
		return nil, err
	}
	return loc, nil
}

func (s *WarehouseLocationService) Delete(ctx context.Context, user *auth.User, warehouseID, id uuid.UUID) error {
	loc, err := s.Get(ctx, user, warehouseID, id)
	if err != nil {
		return err
	}

	// Rak yang masih berisi barang tidak boleh dihapus; pindahkan dulu lewat Stock Transfer.
	stock, err := s.locations.TotalStock(ctx, id)
	if err != nil {
		return err
	}
	if stock > 0 {
		return apperror.Conflict(fmt.Sprintf("lokasi masih berisi %d barang, tidak bisa dihapus", stock))
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.locations.WithTx(tx).Delete(ctx, loc); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionDelete, model.EntityWarehouseLocation,
			loc.ID, fmt.Sprintf("menghapus lokasi %s", loc.Code)))
	})
}

// checkWarehouse: gudang harus ada dan boleh diakses user (gudang orang lain = 404).
func (s *WarehouseLocationService) checkWarehouse(ctx context.Context, user *auth.User, warehouseID uuid.UUID) error {
	_, err := s.warehouses.FindByID(ctx, user.WarehouseScope(), warehouseID)
	if errors.Is(err, repository.ErrNotFound) {
		return apperror.NotFound("warehouse")
	}
	return err
}

func (s *WarehouseLocationService) checkCode(ctx context.Context, warehouseID uuid.UUID, code string, excludeID *uuid.UUID) error {
	exists, err := s.locations.ExistsByCode(ctx, warehouseID, code, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return apperror.Conflict("kode lokasi sudah dipakai di warehouse ini")
	}
	return nil
}
