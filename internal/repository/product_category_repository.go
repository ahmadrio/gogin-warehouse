package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"warehouse-management/internal/model"
	"warehouse-management/pkg/pagination"
)

type ProductCategoryRepository struct {
	db *gorm.DB
}

func NewProductCategoryRepository(db *gorm.DB) *ProductCategoryRepository {
	return &ProductCategoryRepository{db: db}
}

func (r *ProductCategoryRepository) WithTx(tx *gorm.DB) *ProductCategoryRepository {
	return &ProductCategoryRepository{db: tx}
}

func (r *ProductCategoryRepository) List(ctx context.Context, p pagination.Params) ([]model.ProductCategory, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.ProductCategory{})
	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where("code ILIKE ? OR name ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := []model.ProductCategory{}
	err := q.Scopes(paginate(p)).Order("name ASC").Find(&items).Error
	return items, total, err
}

func (r *ProductCategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ProductCategory, error) {
	var c model.ProductCategory
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	return &c, notFound(err)
}

// ExistsByCode: excludeID diisi saat update agar data itu sendiri tidak dihitung duplikat.
func (r *ProductCategoryRepository) ExistsByCode(ctx context.Context, code string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.ProductCategory{}).Where("code = ?", code)
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var count int64
	err := q.Count(&count).Error
	return count > 0, err
}

// CountProducts: jumlah produk (belum dihapus) yang memakai kategori ini.
func (r *ProductCategoryRepository) CountProducts(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Product{}).Where("product_category_id = ?", id).Count(&count).Error
	return count, err
}

func (r *ProductCategoryRepository) Create(ctx context.Context, c *model.ProductCategory) error {
	return r.db.WithContext(ctx).Create(c).Error
}

// Update menyimpan semua kolom (termasuk nilai false/nil).
func (r *ProductCategoryRepository) Update(ctx context.Context, c *model.ProductCategory) error {
	return r.db.WithContext(ctx).Save(c).Error
}

// Delete = soft delete (karena model punya gorm.DeletedAt).
func (r *ProductCategoryRepository) Delete(ctx context.Context, c *model.ProductCategory) error {
	return r.db.WithContext(ctx).Delete(c).Error
}
