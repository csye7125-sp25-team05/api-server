package service

import (
	"api-server/internal/model"
	"api-server/internal/repository"
	"database/sql"

	"github.com/google/uuid"
)

type UserService struct {
	ur *repository.UserRepository
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{ur: repository.NewUserRepository(db)}
}

func (us *UserService) GetAllUsers() ([]model.User, error) {
	return us.ur.GetAllUsers()
}

func (us *UserService) GetUserByID(id uuid.UUID) (*model.User, error) {
	return us.ur.GetUserByID(id)
}

func (us *UserService) CreateUser(user *model.User) error {
	return us.ur.CreateUser(user)
}

func (us *UserService) UpdateUser(id uuid.UUID, user *model.User) error {
	return us.ur.UpdateUser(id, user)
}

func (us *UserService) DeleteUser(id uuid.UUID) error {
	return us.ur.DeleteUser(id)
}
