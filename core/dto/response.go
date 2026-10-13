package dto

import "time"

type SuccessResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type ListResponse struct {
	Status string         `json:"status"`
	Data   interface{}    `json:"data"`
	Meta   PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ErrorResponse struct {
	Status    string `json:"status"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func Success(data interface{}) SuccessResponse {
	return SuccessResponse{Status: "success", Data: data}
}

func List(data interface{}, meta PaginationMeta) ListResponse {
	return ListResponse{Status: "success", Data: data, Meta: meta}
}

func NewPaginationMeta(page, limit, total int) PaginationMeta {
	totalPages := (total + limit - 1) / limit
	return PaginationMeta{Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}

func Error(code, message string) ErrorResponse {
	return ErrorResponse{
		Status:    "error",
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}
