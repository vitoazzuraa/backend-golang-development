package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	NIM       int       `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	OwnerID   *int      `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateStudentRequest struct {
	NIM   int     `json:"nim"   validate:"nim"`
	Name  string  `json:"name"  validate:"required"`
	Grade float64 `json:"grade" validate:"min=0,max=100"`
}

type ReplaceStudentRequest struct {
	NIM      int     `json:"nim"       validate:"nim"`
	Name     string  `json:"name"      validate:"required"`
	Grade    float64 `json:"grade"     validate:"min=0,max=100"`
	IsActive *bool   `json:"is_active" validate:"required"`
}

type PatchStudentRequest struct {
	NIM      *int     `json:"nim,omitempty"   validate:"omitnil,nim"`
	Name     *string  `json:"name,omitempty"  validate:"omitnil,min=1"`
	Grade    *float64 `json:"grade,omitempty" validate:"omitnil,min=0,max=100"`
	IsActive *bool    `json:"is_active,omitempty"`
}

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type Cursor struct {
	CreatedAt time.Time
	ID        int
}

type CursorQuery struct {
	Limit    int
	Search   string
	IsActive *bool
	After    *Cursor
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
