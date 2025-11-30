package models

type ErrorResponse struct {
		Error   string `json:"error"`
    Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
		Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type PaginationMeta struct {
    Page       int   `json:"page"`
    Limit      int   `json:"limit"`
    TotalItems int64 `json:"total_items"`
    TotalPages int   `json:"total_pages"`
}

type PaginatedResponse struct {
    Data       interface{}    `json:"data"`
    Pagination PaginationMeta `json:"pagination"`
}