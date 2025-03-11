package model

import (
	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	AccountCreated string    `json:"account_created"`
	AccountUpdated string    `json:"account_updated"`
}
