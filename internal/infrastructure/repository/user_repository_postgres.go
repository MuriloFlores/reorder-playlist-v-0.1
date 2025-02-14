package repository

import (
	"errors"
	"gorm.io/gorm"
	"project/internal/DTOs"
	"project/internal/core/entities"
	"project/internal/infrastructure/logging"
)

type UserRepositoryInterface interface {
	CreateUser(user entities.UserInterface) error
	GetUserByID(id string) (entities.UserInterface, error)
	GetUserByEmail(email string) (entities.UserInterface, error)
	UpdateUser(user entities.UserInterface) error
	DeleteUser(id string) error
}

type userRepositoryPostgres struct {
	db *gorm.DB
}

func NewUserRepositoryPostgres(db *gorm.DB) UserRepositoryInterface {
	return &userRepositoryPostgres{
		db: db,
	}
}

func (r *userRepositoryPostgres) CreateUser(user entities.UserInterface) error {
	userDTO := DTOs.UserFromEntity(user)
	result := r.db.Create(&userDTO)

	return result.Error
}

func (r *userRepositoryPostgres) GetUserByID(id string) (entities.UserInterface, error) {
	var user DTOs.UserDTO
	result := r.db.First(&user, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		logging.Info("Usuario nao encontrado - user_repository_postgres - GetUserByID - LN40")
		return nil, nil
	}

	return user.ToEntity(), result.Error
}

func (r *userRepositoryPostgres) GetUserByEmail(email string) (entities.UserInterface, error) {
	var user DTOs.UserDTO
	result := r.db.First(&user, "email = ?", email)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return user.ToEntity(), result.Error
}

func (r *userRepositoryPostgres) UpdateUser(user entities.UserInterface) error {
	userDTO := DTOs.UserFromEntity(user)
	result := r.db.Save(&userDTO)
	return result.Error
}

func (r *userRepositoryPostgres) DeleteUser(id string) error {
	result := r.db.Delete(&DTOs.UserDTO{}, "id = ?", id)
	return result.Error

}
