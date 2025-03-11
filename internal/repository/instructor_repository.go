package repository

import (
	"database/sql"
	"errors"

	"api-server/internal/model"

	"github.com/google/uuid"
)

type InstructorRepository struct {
	db *sql.DB
}

func NewInstructorRepository(db *sql.DB) *InstructorRepository {
	return &InstructorRepository{db: db}
}

func (ir *InstructorRepository) GetAllInstructors() ([]model.Instructor, error) {
	rows, err := ir.db.Query(`SELECT * FROM "api"."instructor"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instructors []model.Instructor
	for rows.Next() {
		var instructor model.Instructor
		err = rows.Scan(&instructor.ID, &instructor.UserID, &instructor.Name, &instructor.DateCreated)
		if err != nil {
			return nil, err
		}
		instructors = append(instructors, instructor)
	}
	return instructors, nil
}

func (ir *InstructorRepository) GetInstructorByID(id uuid.UUID) (*model.Instructor, error) {
	row := ir.db.QueryRow("SELECT * FROM api.instructor WHERE id = $1", id)
	var instructor model.Instructor
	err := row.Scan(&instructor.ID, &instructor.UserID, &instructor.Name, &instructor.DateCreated)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("instructor not found")
		}
		return nil, err
	}
	return &instructor, nil
}

func (ir *InstructorRepository) CreateInstructor(instructor *model.Instructor) error {
	if instructor.ID == uuid.Nil {
		instructor.ID = uuid.New()
	}
	_, err := ir.db.Exec("INSERT INTO api.instructor (id, user_id, name) VALUES ($1, $2, $3)",
		instructor.ID, instructor.UserID, instructor.Name)
	return err
}

func (ir *InstructorRepository) UpdateInstructor(id uuid.UUID, instructor *model.Instructor) error {
	_, err := ir.db.Exec("UPDATE api.instructor SET user_id = $1, name = $2 WHERE id = $3",
		instructor.UserID, instructor.Name, id)
	return err
}

func (ir *InstructorRepository) DeleteInstructor(id uuid.UUID) error {
	_, err := ir.db.Exec("DELETE FROM api.instructor WHERE id = $1", id)
	return err
}
