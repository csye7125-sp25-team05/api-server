package model

import (
	"github.com/google/uuid"
)

type Trace struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	FileName    string    `json:"file_name"`
	DateCreated string    `json:"date_created"`
	BucketPath  string    `json:"bucket_path"`
}
