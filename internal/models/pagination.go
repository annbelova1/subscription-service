// internal/models/pagination.go
package models

type Pagination struct {
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}

type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Pagination *Pagination `json:"pagination"`
}