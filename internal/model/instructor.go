package model

import (
	"github.com/google/uuid"
)

type Instructor struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	DateCreated string    `json:"date_created"`
}
