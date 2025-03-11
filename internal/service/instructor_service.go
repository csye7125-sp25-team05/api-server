package service

import (
	"api-server/internal/model"
	"api-server/internal/repository"

	"github.com/google/uuid"
)

type InstructorService struct {
	ir *repository.InstructorRepository
}

func NewInstructorService(ir *repository.InstructorRepository) *InstructorService {
	return &InstructorService{ir: ir}
}

func (is *InstructorService) GetAllInstructors() ([]model.Instructor, error) {
	return is.ir.GetAllInstructors()
}

func (is *InstructorService) GetInstructorByID(id uuid.UUID) (*model.Instructor, error) {
	return is.ir.GetInstructorByID(id)
}

func (is *InstructorService) CreateInstructor(instructor *model.Instructor) error {
	return is.ir.CreateInstructor(instructor)
}

func (is *InstructorService) UpdateInstructor(id uuid.UUID, instructor *model.Instructor) error {
	return is.ir.UpdateInstructor(id, instructor)
}

func (is *InstructorService) DeleteInstructor(id uuid.UUID) error {
	return is.ir.DeleteInstructor(id)
}
