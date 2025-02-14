package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"net/http"
	"project/internal/DTOs"
	"project/internal/core/services"
	"project/internal/infrastructure/logging"
	"project/internal/infrastructure/repository"
	"project/internal/infrastructure/sessions"
)

type getAllPlaylistsHandler struct {
	PlaylistService services.YoutubePlaylistService
	Session         sessions.SessionManager
	repo            repository.UserRepositoryInterface
}

type GetAllPlaylistsHandlerInterface interface {
	GetAllPlaylists(w http.ResponseWriter, r *http.Request)
}

func NewGetAllPlaylistsHandler(playlistService services.YoutubePlaylistService, session sessions.SessionManager, UserRepositoryInterface repository.UserRepositoryInterface) GetAllPlaylistsHandlerInterface {
	return &getAllPlaylistsHandler{
		PlaylistService: playlistService,
		Session:         session,
		repo:            UserRepositoryInterface,
	}
}

func (h *getAllPlaylistsHandler) GetAllPlaylists(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("X-User-Id")
	if userId == "" {
		logging.Error("GetAllPlaylists - header missing", zap.String("header", "X-User-Id"))
		http.Error(w, "Unauthorized: User ID header missing", http.StatusUnauthorized)
		return
	}

	logging.Info("GetAllPlaylists - Get all playlists", zap.String("userId", userId))

	user, err := h.repo.GetUserByID(userId)
	if err != nil {
		logging.Error("GetAllPlaylists - get_all_playlists_handler", zap.String("getting user error", err.Error()))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	oAuthToken := &oauth2.Token{AccessToken: user.Token()}

	playlists, err := h.PlaylistService.GetAllPlaylists(r.Context(), oAuthToken, r)
	if err != nil {
		logging.Error("GetAllPlaylists", zap.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	playlistsDTO := make([]DTOs.PlaylistRedisDTO, len(playlists))
	for i, playlist := range playlists {
		playlistsDTO[i] = DTOs.PlaylistFromEntity(playlist)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(playlistsDTO); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
