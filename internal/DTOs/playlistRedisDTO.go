package DTOs

import (
	"project/internal/core/entities"
	"time"
)

type PlaylistRedisDTO struct {
	Id          string          `json:"id"`
	ChannelId   string          `json:"channelId"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	PublishedAt time.Time       `json:"publishedAt"`
	Videos      []VideoRedisDTO `json:"videos"`
}

func (dto *PlaylistRedisDTO) ToEntity() entities.PlaylistInterface {

	videos := make([]entities.VideoInterface, len(dto.Videos))

	for i, video := range dto.Videos {
		videos[i] = video.ToEntity()
	}

	return entities.NewPlaylist(
		dto.Id,
		dto.ChannelId,
		dto.Title,
		dto.Description,
		dto.PublishedAt,
		videos,
	)
}

func PlaylistFromEntity(entity entities.PlaylistInterface) PlaylistRedisDTO {
	videos := make([]VideoRedisDTO, len(entity.Videos()))
	for i, video := range entity.Videos() {
		videos[i] = VideoFromEntity(video)
	}

	return PlaylistRedisDTO{
		Id:          entity.Id(),
		ChannelId:   entity.ChannelId(),
		Title:       entity.Title(),
		Description: entity.Description(),
		PublishedAt: entity.PublishedAt(),
		Videos:      videos,
	}
}
