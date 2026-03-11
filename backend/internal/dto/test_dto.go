package dto

import "time"

type TestCreateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type TestResponse struct {
	ID uint `json:"id"`
	Name string `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TestListResponse struct {
	Data []TestResponse `json:"data"`
}