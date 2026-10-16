// Package handler: lapisan HTTP (padanan Controller di Laravel).
package handler

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"warehouse-management/pkg/apperror"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// bindJSON membaca body JSON ke struct dan menjalankan validasi dari tag `binding`.
// Error validasi diubah ke format {"field": "pesan"} dengan status 422.
func bindJSON(c *gin.Context, req any) error {
	if err := c.ShouldBindJSON(req); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			details := make(map[string]string, len(verrs))
			for _, fe := range verrs {
				details[fieldPath(fe)] = validationMessage(fe)
			}
			return apperror.Validation(details)
		}
		return apperror.BadRequest("format JSON tidak valid")
	}
	return nil
}

// paramUUID membaca path parameter (mis. /:id) sebagai UUID.
func paramUUID(c *gin.Context, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return uuid.Nil, apperror.BadRequest(name + " harus berupa UUID")
	}
	return id, nil
}

func validationMessage(fe validator.FieldError) string {
	isList := fe.Kind() == reflect.Slice || fe.Kind() == reflect.Map
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		if isList {
			return fmt.Sprintf("minimal %s item", fe.Param())
		}
		return fmt.Sprintf("minimal %s karakter", fe.Param())
	case "max":
		if isList {
			return fmt.Sprintf("maksimal %s item", fe.Param())
		}
		return fmt.Sprintf("maksimal %s karakter", fe.Param())
	case "ne":
		return "tidak boleh " + fe.Param()
	case "gte":
		return fmt.Sprintf("minimal %s", fe.Param())
	case "uuid":
		return "harus berupa UUID"
	case "oneof":
		return "harus salah satu dari: " + fe.Param()
	default:
		return "tidak valid (" + fe.Tag() + ")"
	}
}

// toSnakeCase: "ProductCategoryID" -> "product_category_id"
func toSnakeCase(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		isUpper := r >= 'A' && r <= 'Z'
		if isUpper && i > 0 {
			prevLower := runes[i-1] >= 'a' && runes[i-1] <= 'z'
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			if prevLower || nextLower {
				b.WriteByte('_')
			}
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

// queryUUID membaca query string opsional (mis. ?warehouse_id=) sebagai UUID.
// Tidak dikirim -> nil (tidak memfilter). Dikirim tapi bukan UUID -> 400.
func queryUUID(c *gin.Context, name string) (*uuid.UUID, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, apperror.BadRequest(name + " harus berupa UUID")
	}
	return &id, nil
}

// fieldPath membentuk nama field untuk pesan error, termasuk field di dalam array:
// "StockAdjustmentRequest.Items[0].QuantityChange" -> "items[0].quantity_change"
func fieldPath(fe validator.FieldError) string {
	parts := strings.Split(fe.Namespace(), ".")
	if len(parts) > 1 {
		parts = parts[1:] // buang nama struct paling luar
	}
	for i, part := range parts {
		parts[i] = toSnakeCase(part)
	}
	return strings.Join(parts, ".")
}
