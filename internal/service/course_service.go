package service

import (
	"api-server/internal/model"
	"api-server/internal/repository"
	"database/sql"

	"github.com/google/uuid"
)

type CourseService struct {
	cr *repository.CourseRepository
}

func NewCourseService(db *sql.DB) *CourseService {
	return &CourseService{cr: repository.NewCourseRepository(db)}
}

func (cs *CourseService) GetAllCourses() ([]model.Course, error) {
	return cs.cr.GetAllCourses()
}

func (cs *CourseService) GetCourseByID(id uuid.UUID) (*model.Course, error) {
	return cs.cr.GetCourseByID(id)
}

func (cs *CourseService) CreateCourse(course *model.Course) error {
	return cs.cr.CreateCourse(course)
}

func (cs *CourseService) UpdateCourse(id uuid.UUID, course *model.Course) error {
	return cs.cr.UpdateCourse(id, course)
}

func (cs *CourseService) DeleteCourse(id uuid.UUID) error {
	return cs.cr.DeleteCourse(id)
}
