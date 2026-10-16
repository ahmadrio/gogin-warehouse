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

type ProductService struct {
	db         *gorm.DB
	products   *repository.ProductRepository
	categories *repository.ProductCategoryRepository
	audits     *repository.AuditRepository
}

func NewProductService(db *gorm.DB, products *repository.ProductRepository, categories *repository.ProductCategoryRepository, audits *repository.AuditRepository) *ProductService {
	return &ProductService{db: db, products: products, categories: categories, audits: audits}
}

func (s *ProductService) List(ctx context.Context, f repository.ProductFilter, p pagination.Params) ([]model.Product, int64, error) {
	return s.products.List(ctx, f, p)
}

func (s *ProductService) Get(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	p, err := s.products.FindByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("produk")
	}
	return p, err
}

func (s *ProductService) Create(ctx context.Context, user *auth.User, req dto.CreateProductRequest) (*model.Product, error) {
	if err := s.validate(ctx, req.ProductCategoryID, req.SKU, nil); err != nil {
		return nil, err
	}

	product := &model.Product{
		ProductCategoryID: req.ProductCategoryID,
		SKU:               req.SKU,
		Name:              req.Name,
		Description:       req.Description,
		Unit:              req.Unit,
		MinimumStock:      req.MinimumStock,
		IsActive:          req.IsActive == nil || *req.IsActive,
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.products.WithTx(tx).Create(ctx, product); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionCreate, model.EntityProduct,
			product.ID, fmt.Sprintf("membuat produk %s", product.SKU)))
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, apperror.Conflict("SKU sudah dipakai")
	}
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, product.ID) // baca ulang agar kategori ikut ter-preload
}

func (s *ProductService) Update(ctx context.Context, user *auth.User, id uuid.UUID, req dto.UpdateProductRequest) (*model.Product, error) {
	product, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.validate(ctx, req.ProductCategoryID, req.SKU, &id); err != nil {
		return nil, err
	}

	product.ProductCategoryID = req.ProductCategoryID
	product.SKU = req.SKU
	product.Name = req.Name
	product.Description = req.Description
	product.Unit = req.Unit
	product.MinimumStock = req.MinimumStock
	product.IsActive = *req.IsActive

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.products.WithTx(tx).Update(ctx, product); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionUpdate, model.EntityProduct,
			product.ID, fmt.Sprintf("mengubah produk %s", product.SKU)))
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, apperror.Conflict("SKU sudah dipakai")
	}
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *ProductService) Delete(ctx context.Context, user *auth.User, id uuid.UUID) error {
	product, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Produk yang masih punya stok tidak boleh dihapus; kosongkan dulu lewat transaksi.
	stock, err := s.products.TotalStock(ctx, id)
	if err != nil {
		return err
	}
	if stock > 0 {
		return apperror.Conflict(fmt.Sprintf("produk masih memiliki stok %d, tidak bisa dihapus", stock))
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.products.WithTx(tx).Delete(ctx, product); err != nil {
			return err
		}
		return s.audits.WithTx(tx).Create(ctx, newAudit(user, model.ActionDelete, model.EntityProduct,
			product.ID, fmt.Sprintf("menghapus produk %s", product.SKU)))
	})
}

// validate: kategori harus ada dan aktif, SKU harus unik.
func (s *ProductService) validate(ctx context.Context, categoryID uuid.UUID, sku string, excludeID *uuid.UUID) error {
	category, err := s.categories.FindByID(ctx, categoryID)
	if errors.Is(err, repository.ErrNotFound) || (err == nil && !category.IsActive) {
		return apperror.Validation(map[string]string{"product_category_id": "kategori tidak ditemukan atau tidak aktif"})
	}
	if err != nil {
		return err
	}

	exists, err := s.products.ExistsBySKU(ctx, sku, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return apperror.Conflict("SKU sudah dipakai")
	}
	return nil
}
