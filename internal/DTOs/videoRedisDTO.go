package DTOs

import (
	"project/internal/core/entities"
	"time"
)

type VideoRedisDTO struct {
	Id          string        `json:"id"`
	Title       string        `json:"title"`
	ChannelId   string        `json:"artist"`
	Language    string        `json:"language"`
	PublishedAt time.Time     `json:"published_at"`
	Duration    time.Duration `json:"duration"`
}

func (dto *VideoRedisDTO) ToEntity() entities.VideoInterface {
	return entities.NewVideo(
		dto.Id,
		dto.Title,
		dto.ChannelId,
		dto.Language,
		dto.PublishedAt,
		dto.Duration,
	)
}

func VideoFromEntity(entity entities.VideoInterface) VideoRedisDTO {
	return VideoRedisDTO{
		Id:          entity.Id(),
		Title:       entity.Title(),
		ChannelId:   entity.ChannelId(),
		Language:    entity.Language(),
		PublishedAt: entity.PublishedAt(),
		Duration:    entity.Duration(),
	}
}
