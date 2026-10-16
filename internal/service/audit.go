package service

import (
	"time"

	"github.com/google/uuid"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/model"
)

// newAudit membentuk satu baris audit trail.
// Audit selalu ditulis di DALAM transaksi yang sama dengan perubahan datanya.
func newAudit(user *auth.User, action, entityType string, entityID uuid.UUID, description string) *model.AuditTrail {
	var desc *string
	if description != "" {
		desc = &description
	}
	return &model.AuditTrail{
		UserID:      user.ID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: desc,
		OccurredAt:  time.Now(),
	}
}
