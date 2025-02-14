package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"project/internal/core/usecases"
	"project/internal/infrastructure/logging"
	"project/internal/infrastructure/sessions"
)

type reorderPlaylistHandler struct {
	ReorderUseCase usecases.ReorderPlaylistUseCaseInterface
	Session        sessions.SessionManager
}

type ReorderPlaylistHandlerInterface interface {
	ReorderPlaylist(w http.ResponseWriter, r *http.Request)
}

func NewPlaylistHandler(uc usecases.ReorderPlaylistUseCaseInterface, session sessions.SessionManager) ReorderPlaylistHandlerInterface {
	return &reorderPlaylistHandler{
		ReorderUseCase: uc,
		Session:        session,
	}
}

type ReorderPlaylistRequest struct {
	PlaylistId string `json:"playlist_id"`
	Criteria   string `json:"criteria"` // Ex.: "byName", "byPublishedAt", "byDuration"
}

func (h *reorderPlaylistHandler) ReorderPlaylist(w http.ResponseWriter, r *http.Request) {
	var req ReorderPlaylistRequest
	userId := h.Session.GetUserId(r)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Requisition", http.StatusBadRequest)
		return
	}
	if req.PlaylistId == "" || req.Criteria == "" {
		http.Error(w, "playlist_id e criteria são obrigatórios", http.StatusBadRequest)
		return
	}
	if err := h.ReorderUseCase.Execute(req.PlaylistId, req.Criteria, userId, r.Context()); err != nil {
		logging.Error("Erro ao reordenar playlist - reorder_playlist_handler - ln46", zap.String("err", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]string{"message": "Playlist reordenada com sucesso"}
	logging.Info("Reordenada com sucesso", zap.String("ID: "+req.PlaylistId, "Criteria: "+req.Criteria))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
