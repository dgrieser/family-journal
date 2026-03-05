package services

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PaginationParams struct {
	Page     int
	PageSize int
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

type PaginatedResponse[T any] struct {
	Items      []T            `json:"items"`
	Pagination PaginationMeta `json:"pagination"`
}

func NewPagination(page, pageSize int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func NewPaginatedResponse[T any](items []T, totalItems int, params PaginationParams) PaginatedResponse[T] {
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + params.PageSize - 1) / params.PageSize
	}
	return PaginatedResponse[T]{
		Items: items,
		Pagination: PaginationMeta{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	}
}
