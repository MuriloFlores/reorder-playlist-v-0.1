package DTOs

import (
	"project/internal/core/entities"
	"time"
)

type VideoRedisDTO struct {
	Id          string        `json:"id"`
	Title       string        `json:"title"`
	Artist      string        `json:"artist"`
	PublishedAt time.Time     `json:"published_at"`
	Duration    time.Duration `json:"duration"`
}

func (dto *VideoRedisDTO) ToEntity() entities.VideoInterface {
	return entities.NewVideo(
		dto.Id,
		dto.Title,
		dto.Artist,
		dto.PublishedAt,
		dto.Duration,
	)
}

func VideoFromEntity(entity entities.VideoInterface) VideoRedisDTO {
	return VideoRedisDTO{
		Id:          entity.Id(),
		Title:       entity.Title(),
		Artist:      entity.Artist(),
		PublishedAt: entity.PublishedAt(),
		Duration:    entity.Duration(),
	}
}
