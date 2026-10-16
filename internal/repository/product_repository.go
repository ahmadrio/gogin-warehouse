package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"warehouse-management/internal/model"
	"warehouse-management/pkg/pagination"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) WithTx(tx *gorm.DB) *ProductRepository {
	return &ProductRepository{db: tx}
}

type ProductFilter struct {
	CategoryID *uuid.UUID
}

func (r *ProductRepository) List(ctx context.Context, f ProductFilter, p pagination.Params) ([]model.Product, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Product{})
	if f.CategoryID != nil {
		q = q.Where("product_category_id = ?", *f.CategoryID)
	}
	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where("sku ILIKE ? OR name ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := []model.Product{}
	err := q.Preload("ProductCategory").Scopes(paginate(p)).Order("sku ASC").Find(&items).Error
	return items, total, err
}

func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	var p model.Product
	err := r.db.WithContext(ctx).Preload("ProductCategory").First(&p, "id = ?", id).Error
	return &p, notFound(err)
}

// ExistsBySKU: excludeID diisi saat update agar produk itu sendiri tidak dihitung duplikat.
func (r *ProductRepository) ExistsBySKU(ctx context.Context, sku string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Product{}).Where("sku = ?", sku)
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var count int64
	err := q.Count(&count).Error
	return count > 0, err
}

// TotalStock: jumlah stok produk ini di semua gudang dan lokasi.
func (r *ProductRepository) TotalStock(ctx context.Context, id uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.WarehouseLocationItem{}).
		Where("product_id = ?", id).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total).Error
	return total, err
}

func (r *ProductRepository) Create(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(p).Error
}

// Update menyimpan semua kolom produk. Omit(clause.Associations) mencegah GORM
// ikut menyimpan ProductCategory yang ter-preload.
func (r *ProductRepository) Update(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Save(p).Error
}

func (r *ProductRepository) Delete(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Delete(p).Error
}
