package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SequenceRepository membuat nomor dokumen unik (BR-003), contoh ADJ-202610-0001.
type SequenceRepository struct {
	db *gorm.DB
}

func NewSequenceRepository(db *gorm.DB) *SequenceRepository {
	return &SequenceRepository{db: db}
}

func (r *SequenceRepository) WithTx(tx *gorm.DB) *SequenceRepository {
	return &SequenceRepository{db: tx}
}

// Next menaikkan counter (prefix, bulan ini) secara atomik lalu membentuk nomornya.
func (r *SequenceRepository) Next(ctx context.Context, prefix string, now time.Time) (string, error) {
	period := now.Format("200601") // format tanggal Go: 2006=tahun, 01=bulan -> "202610"

	var last int
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO document_sequences (prefix, period, last_number) VALUES (?, ?, 1)
		ON CONFLICT (prefix, period)
		DO UPDATE SET last_number = document_sequences.last_number + 1, updated_at = now()
		RETURNING last_number`, prefix, period).Scan(&last).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%04d", prefix, period, last), nil
}
