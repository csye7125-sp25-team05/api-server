package model

import (
	"github.com/google/uuid"
)

type Course struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	SemesterTerm    string    `json:"semester_term"`
	Manufacturer    string    `json:"manufacturer"`
	CreditHours     int       `json:"credit_hours"`
	SemesterYear    int       `json:"semester_year"`
	DateAdded       string    `json:"date_added"`
	DateLastUpdated string    `json:"date_last_updated"`
	OwnerUserID     uuid.UUID `json:"owner_user_id"`
	InstructorID    uuid.UUID `json:"instructor_id"`
}
