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

type ProductCategoryService struct {
	db         *gorm.DB // dibutuhkan untuk membuka transaksi
	categories *repository.ProductCategoryRepository
	audits     *repository.AuditRepository
}

func NewProductCategoryService(db *gorm.DB, categories *repository.ProductCategoryRepository, audits *repository.AuditRepository) *ProductCategoryService {
	return &ProductCategoryService{db: db, categories: categories, audits: audits}
}

func (s *ProductCategoryService) List(ctx context.Context, p pagination.Params) ([]model.ProductCategory, int64, error) {
	return s.categories.List(ctx, p)
}

func (s *ProductCategoryService) Get(ctx context.Context, id uuid.UUID) (*model.ProductCategory, error) {
	c, err := s.categories.FindByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("kategori produk")
	}
	return c, err
}

func (s *ProductCategoryService) Create(ctx context.Context, user *auth.User, req dto.CreateProductCategoryRequest) (*model.ProductCategory, error) {
	// 1. Validasi bisnis
	exists, err := s.categories.ExistsByCode(ctx, req.Code, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.Conflict("kode kategori sudah dipakai")
	}

	category := &model.ProductCategory{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive == nil || *req.IsActive,
	}

	// 2. Simpan data + audit trail dalam satu transaksi
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.categories.WithTx(tx).Create(ctx, category); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionCreate, model.EntityProductCategory,
			category.ID, fmt.Sprintf("membuat kategori %s", category.Code)))
	})
	// 3. Pengaman terakhir: dua request bersamaan lolos ExistsByCode, unique index DB yang menolak
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, apperror.Conflict("kode kategori sudah dipakai")
	}
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (s *ProductCategoryService) Update(ctx context.Context, user *auth.User, id uuid.UUID, req dto.UpdateProductCategoryRequest) (*model.ProductCategory, error) {
	category, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	exists, err := s.categories.ExistsByCode(ctx, req.Code, &id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.Conflict("kode kategori sudah dipakai")
	}

	category.Code = req.Code
	category.Name = req.Name
	category.Description = req.Description
	category.IsActive = *req.IsActive

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.categories.WithTx(tx).Update(ctx, category); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionUpdate, model.EntityProductCategory,
			category.ID, fmt.Sprintf("mengubah kategori %s", category.Code)))
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, apperror.Conflict("kode kategori sudah dipakai")
	}
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (s *ProductCategoryService) Delete(ctx context.Context, user *auth.User, id uuid.UUID) error {
	category, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Aturan bisnis: kategori yang masih dipakai produk tidak boleh dihapus.
	used, err := s.categories.CountProducts(ctx, id)
	if err != nil {
		return err
	}
	if used > 0 {
		return apperror.Conflict(fmt.Sprintf("kategori masih dipakai oleh %d produk", used))
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.categories.WithTx(tx).Delete(ctx, category); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionDelete, model.EntityProductCategory,
			category.ID, fmt.Sprintf("menghapus kategori %s", category.Code)))
	})
}
