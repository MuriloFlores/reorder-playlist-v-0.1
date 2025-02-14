package usecases

import (
	"context"
	"project/internal/core/services"
)

type reorderPlaylistUseCase struct {
	PlaylistService services.YoutubePlaylistService
}

type ReorderPlaylistUseCaseInterface interface {
	Execute(playlistId, criteria, userId string, ctx context.Context) error
}

func NewReorderPlaylistUseCase(service services.YoutubePlaylistService) ReorderPlaylistUseCaseInterface {
	return &reorderPlaylistUseCase{
		PlaylistService: service,
	}
}

func (uc *reorderPlaylistUseCase) Execute(playlistId, criteria, userId string, ctx context.Context) error {
	return uc.PlaylistService.ReorderPlaylist(playlistId, criteria, userId, ctx)
}
