// Package pagination membaca ?page=&per_page=&search= dan membentuk meta pagination.
package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPerPage = 15
	maxPerPage     = 100
)

type Params struct {
	Page    int
	PerPage int
	Search  string
}

func (p Params) Offset() int { return (p.Page - 1) * p.PerPage }

func FromQuery(c *gin.Context) Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(defaultPerPage)))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > maxPerPage {
		perPage = defaultPerPage
	}
	return Params{Page: page, PerPage: perPage, Search: c.Query("search")}
}

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func NewMeta(p Params, total int64) Meta {
	pages := total / int64(p.PerPage)
	if total%int64(p.PerPage) != 0 {
		pages++
	}
	return Meta{Page: p.Page, PerPage: p.PerPage, Total: total, TotalPages: pages}
}
