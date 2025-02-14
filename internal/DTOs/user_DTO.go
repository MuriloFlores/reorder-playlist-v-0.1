package DTOs

import (
	"project/internal/core/entities"
	"time"
)

type UserDTO struct {
	ID           string    `json:"id" gorm:"primaryKey;column:id"`
	Name         string    `json:"name" gorm:"column:name"`
	Email        string    `json:"email" gorm:"column:email"`
	RefreshToken string    `json:"refresh_token;omitempty" gorm:"column:refresh_token;not null"`
	Token        string    `json:"token" gorm:"column:token;not null"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"column:expires_at;type:timestamp"`
}

func (dto *UserDTO) TableName() string {
	return "users"
}

func (dto *UserDTO) ToEntity() entities.UserInterface {
	return entities.NewUser(
		dto.ID,
		dto.Name,
		dto.Email,
		dto.Token,
		dto.RefreshToken,
		dto.ExpiresAt,
	)
}

func UserFromEntity(entity entities.UserInterface) UserDTO {
	return UserDTO{
		ID:           entity.Id(),
		Name:         entity.Name(),
		Email:        entity.Email(),
		RefreshToken: entity.RefreshToken(),
		Token:        entity.Token(),
		ExpiresAt:    entity.ExpiresAt(),
	}
}
