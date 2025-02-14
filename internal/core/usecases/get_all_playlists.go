package usecases

import (
	"context"
	"golang.org/x/oauth2"
	"net/http"
	"project/internal/core/entities"
	"project/internal/core/services"
)

type getAllPlaylistUseCase struct {
	PlaylistService services.YoutubePlaylistService
}
type GetAllPlaylistsUseCase interface {
	Execute(ctx context.Context, token *oauth2.Token, r *http.Request) ([]entities.PlaylistInterface, error)
}

func NewGetAllPlaylistsUseCase(service services.YoutubePlaylistService) GetAllPlaylistsUseCase {
	return &getAllPlaylistUseCase{
		PlaylistService: service,
	}
}

func (uc *getAllPlaylistUseCase) Execute(ctx context.Context, token *oauth2.Token, r *http.Request) ([]entities.PlaylistInterface, error) {
	return uc.PlaylistService.GetAllPlaylists(ctx, token, r)
}
