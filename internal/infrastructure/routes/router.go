package routes

import (
	"encoding/json"
	"net/http"
	"project/internal/infrastructure/repository"

	handlers2 "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"project/internal/adapters/handlers"
	"project/internal/core/services"
	"project/internal/infrastructure/middleware"
	"project/internal/infrastructure/sessions"
)

// ConfigureRoutes configura todas as rotas da API.
// Note que, para utilizar o middleware, precisamos injetar também
// o SessionManager e o AuthService, que são usados para validar o token.
func ConfigureRoutes(
	authHandler handlers.AuthHandler,
	reorder handlers.ReorderPlaylistHandlerInterface,
	getAll handlers.GetAllPlaylistsHandlerInterface,
	store sessions.SessionManager,
	authService services.AuthService,
	repo repository.UserRepositoryInterface,
) http.Handler {
	// Cria o router principal
	router := mux.NewRouter()

	// Rotas de autenticação (normalmente públicas)
	router.HandleFunc("/auth/{provider}", authHandler.OAuthLogin).Methods("GET")
	router.HandleFunc("/auth/google/callback", authHandler.OAuthCallback).Methods("GET")
	router.HandleFunc("/auth/logout", authHandler.OAuthLogout).Methods("GET")

	protected := router.PathPrefix("/playlists").Subrouter()
	authMiddleware := middleware.NewAuthMiddleware(store, authService, repo)
	protected.Use(authMiddleware.ValidateTokenHandler)

	protected.HandleFunc("/reorder", reorder.ReorderPlaylist).Methods("POST")
	protected.HandleFunc("/all", getAll.GetAllPlaylists).Methods("GET")
	protected.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"valid": true})
	}).Methods("GET")

	// Configuração de CORS
	corsOption := handlers2.AllowedOrigins([]string{"http://localhost:5173"})
	corsMethods := handlers2.AllowedMethods([]string{"GET", "POST", "OPTIONS", "PUT", "DELETE"})
	corsHeaders := handlers2.AllowedHeaders([]string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-User-Id"})
	corsCredentials := handlers2.AllowCredentials()
	corsHandler := handlers2.CORS(corsOption, corsMethods, corsHeaders, corsCredentials)(router)

	return corsHandler
}
