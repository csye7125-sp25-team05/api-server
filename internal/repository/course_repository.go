package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"api-server/internal/model"

	"github.com/google/uuid"
)

type CourseRepository struct {
	db *sql.DB
}

func NewCourseRepository(db *sql.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (cr *CourseRepository) GetAllCourses() ([]model.Course, error) {
	rows, err := cr.db.Query(`SELECT * FROM "api"."course"`)
	if err != nil {
		fmt.Println("ERROR", err)
		return nil, err
	}
	defer rows.Close()

	var courses []model.Course
	for rows.Next() {
		var course model.Course
		err = rows.Scan(&course.ID, &course.Code, &course.Name, &course.Description, &course.SemesterTerm, &course.Manufacturer, &course.CreditHours, &course.SemesterYear, &course.DateAdded, &course.DateLastUpdated, &course.OwnerUserID, &course.InstructorID)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	return courses, nil
}

func (cr *CourseRepository) GetCourseByID(id uuid.UUID) (*model.Course, error) {
	row := cr.db.QueryRow("SELECT * FROM api.course WHERE id = $1", id)
	var course model.Course
	err := row.Scan(&course.ID, &course.Code, &course.Name, &course.Description, &course.SemesterTerm, &course.Manufacturer, &course.CreditHours, &course.SemesterYear, &course.DateAdded, &course.DateLastUpdated, &course.OwnerUserID, &course.InstructorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("course not found")
		}
		return nil, err
	}
	return &course, nil
}

func (cr *CourseRepository) CreateCourse(course *model.Course) error {
	if course.ID == uuid.Nil {
		course.ID = uuid.New()
	}
	_, err := cr.db.Exec("INSERT INTO api.course (id, code, name, description, semesterterm, manufacturer, credithours, semesteryear, date_added, date_last_updated, owner_user_id, instructorid) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $9, $10)",
		course.ID, course.Code, course.Name, course.Description, course.SemesterTerm, course.Manufacturer, course.CreditHours, course.SemesterYear, course.OwnerUserID, course.InstructorID)
	return err
}

func (cr *CourseRepository) UpdateCourse(id uuid.UUID, course *model.Course) error {
	_, err := cr.db.Exec("UPDATE api.course SET code = $1, name = $2, description = $3, semesterterm = $4, manufacturer = $5, credithours = $6, semesteryear = $7, date_last_updated = CURRENT_TIMESTAMP WHERE id = $8",
		course.Code, course.Name, course.Description, course.SemesterTerm, course.Manufacturer, course.CreditHours, course.SemesterYear, id)
	return err
}

func (cr *CourseRepository) DeleteCourse(id uuid.UUID) error {
	_, err := cr.db.Exec("DELETE FROM api.course WHERE id = $1", id)
	return err
}
