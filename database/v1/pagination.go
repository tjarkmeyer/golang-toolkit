package database

import (
	"math"

	"gorm.io/gorm"
)

type Pagination struct {
	PerPage     int         `json:"perPage,omitempty;query:limit"`
	CurrentPage int         `json:"currentPage,omitempty;query:page"`
	Sort        string      `json:"sort,omitempty;query:sort"`
	Total       int64       `json:"total"`
	TotalPages  int         `json:"totalPages"`
	Data        interface{} `json:"data"`
}

func (p *Pagination) GetOffset() int {
	return (p.GetCurrentPage() - 1) * p.GetPerPage()
}

func (p *Pagination) GetPerPage() int {
	if p.PerPage == 0 {
		p.PerPage = 10
	}
	return p.PerPage
}

func (p *Pagination) GetCurrentPage() int {
	if p.CurrentPage == 0 {
		p.CurrentPage = 1
	}
	return p.CurrentPage
}

func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		p.Sort = "Id desc"
	}
	return p.Sort
}

func Paginate(value interface{}, pagination *Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var total int64
	db.Model(value).Count(&total)

	pagination.Total = total
	totalPages := int(math.Ceil(float64(total) / float64(pagination.GetPerPage())))
	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetPerPage()).Order(pagination.GetSort())
	}
}

func PaginateQuery(value interface{}, pagination *Pagination, db *gorm.DB, query string, args ...interface{}) func(db *gorm.DB) *gorm.DB {
	var total int64
	db.Model(value).Where(query, args...).Count(&total)

	pagination.Total = total
	totalPages := int(math.Ceil(float64(total) / float64(pagination.GetPerPage())))
	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...).Offset(pagination.GetOffset()).Limit(pagination.GetPerPage()).Order(pagination.GetSort())
	}
}
