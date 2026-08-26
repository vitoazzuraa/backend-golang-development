package main

type Student struct {
	ID       int     `json:"id"`
	NIM      int     `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

type CreateStudentRequest struct {
	NIM   int     `json:"nim"`
	Name  string  `json:"name"`
	Grade float64 `json:"grade"`
}

type ReplaceStudentRequest struct {
	NIM      int     `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive *bool   `json:"is_active"`
}

type PatchStudentRequest struct {
	NIM      *int     `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
