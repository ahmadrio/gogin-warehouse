package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/repository"
	"warehouse-management/pkg/apperror"
)

// Helper bersama untuk In-Bound, Out-Bound, dan Stock Transfer.
// (Stock Adjustment di Tahap 13 menulis versinya sendiri sebelum helper ini ada.)

// checkWarehouseAccess: warehouse_id wajib diisi dan harus gudang milik user (BR-002).
func checkWarehouseAccess(ctx context.Context, warehouses *repository.WarehouseRepository, user *auth.User, warehouseID uuid.UUID) error {
	if warehouseID == uuid.Nil {
		return apperror.Validation(map[string]string{"warehouse_id": "wajib diisi"})
	}
	_, err := warehouses.FindByID(ctx, user.WarehouseScope(), warehouseID)
	if errors.Is(err, repository.ErrNotFound) {
		return apperror.NotFound("warehouse")
	}
	return err
}

// masterData menampung produk & lokasi aktif yang dirujuk item transaksi,
// diambil sekaligus (satu query per tabel), lalu dicek per item lewat map.
type masterData struct {
	products  map[uuid.UUID]bool
	locations map[uuid.UUID]string // id -> kode rak, untuk label pesan error
	skus      map[uuid.UUID]string // id -> SKU
}

func loadMasterData(ctx context.Context, lookups *repository.LookupRepository, warehouseID uuid.UUID, productIDs, locationIDs []uuid.UUID) (*masterData, error) {
	products, err := lookups.ActiveProducts(ctx, productIDs)
	if err != nil {
		return nil, err
	}
	m := &masterData{products: map[uuid.UUID]bool{}, locations: map[uuid.UUID]string{}, skus: map[uuid.UUID]string{}}
	for id, p := range products {
		m.products[id] = true
		m.skus[id] = p.SKU
	}
	if len(locationIDs) > 0 {
		locations, err := lookups.ActiveLocations(ctx, warehouseID, locationIDs)
		if err != nil {
			return nil, err
		}
		for id, l := range locations {
			m.locations[id] = l.Code
		}
	}
	return m, nil
}

func (m *masterData) checkProduct(errs map[string]string, field string, id uuid.UUID) {
	if !m.products[id] {
		errs[field] = "produk tidak ditemukan atau tidak aktif"
	}
}

func (m *masterData) checkLocation(errs map[string]string, field string, id uuid.UUID) {
	if _, ok := m.locations[id]; !ok {
		errs[field] = "lokasi tidak ditemukan, tidak aktif, atau bukan milik warehouse ini"
	}
}
