package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"api-server/internal/model"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) GetAllUsers() ([]model.User, error) {
	rows, err := ur.db.Query(`SELECT * FROM "api"."user"`)
	if err != nil {
		fmt.Println("ERROR", err)
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err = rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Password, &user.AccountCreated, &user.AccountUpdated)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (ur *UserRepository) GetUserByID(id uuid.UUID) (*model.User, error) {
	row := ur.db.QueryRow("SELECT * FROM api.user WHERE id = $1", id)
	var user model.User
	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Password, &user.AccountCreated, &user.AccountUpdated)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) CreateUser(user *model.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	_, err := ur.db.Exec("INSERT INTO api.user (id, first_name, last_name, username, password, account_created, account_updated) VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		user.ID, user.FirstName, user.LastName, user.Username, user.Password)
	return err
}

func (ur *UserRepository) UpdateUser(id uuid.UUID, user *model.User) error {
	_, err := ur.db.Exec("UPDATE api.user SET first_name = $1, last_name = $2, username = $3, password = $4, account_updated = CURRENT_TIMESTAMP WHERE id = $5",
		user.FirstName, user.LastName, user.Username, user.Password, id)
	return err
}

func (ur *UserRepository) DeleteUser(id uuid.UUID) error {
	_, err := ur.db.Exec("DELETE FROM api.user WHERE id = $1", id)
	return err
}
