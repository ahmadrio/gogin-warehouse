package repository

import (
	"context"

	"gorm.io/gorm"

	"warehouse-management/internal/model"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// WithTx mengembalikan salinan repository yang memakai transaksi tx.
func (r *AuditRepository) WithTx(tx *gorm.DB) *AuditRepository {
	return &AuditRepository{db: tx}
}

func (r *AuditRepository) Create(ctx context.Context, a *model.AuditTrail) error {
	return r.db.WithContext(ctx).Create(a).Error
}
