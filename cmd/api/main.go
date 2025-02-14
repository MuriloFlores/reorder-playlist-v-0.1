package main

import (
	"log"
	"net/http"
	"project/internal/adapters/handlers"
	"project/internal/core/services"
	"project/internal/core/usecases"
	"project/internal/infrastructure/auth"
	"project/internal/infrastructure/cache"
	"project/internal/infrastructure/config"
	"project/internal/infrastructure/error_handler"
	"project/internal/infrastructure/messaging"
	"project/internal/infrastructure/repository"
	"project/internal/infrastructure/routes"
	"project/internal/infrastructure/sessions"
)

func init() {
	config.InitEnvConfig()

	auth.InitGoth(auth.OAuthConfig{
		ClientID:     config.EnvConfigs.ClientID,
		ClientSecret: config.EnvConfigs.SecretKey,
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/youtube",
			"https://www.googleapis.com/auth/youtube.force-ssl",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/youtubepartner",
			"openid",
			"email",
			"profile",
		},
	})
}

func main() {
	dbConn := config.ConnectDB()

	// Gerenciador de sessões
	sessionManager := sessions.NewGorillaSessionManager(config.EnvConfigs.SessionName, config.EnvConfigs.SessionSecret)
	userRepository := repository.NewUserRepositoryPostgres(dbConn)

	// Serviço de autenticação
	authService := &services.GothAuthService{}
	authHandler := handlers.NewHandler(authService, sessionManager, userRepository)

	// Instancie o cache e o repositório
	redisCache := cache.NewRedisCache("localhost:6379")
	repo := repository.NewPlaylistRepositoryRedis(redisCache)

	producer := messaging.NewRabbitMQProducer("reorderApi")

	// Instancie o tratador de erros (via inversão de dependência)
	errHandler := error_handler.NewErrorHandler(producer)
	// Serviço de playlists
	youtubeService := services.NewYoutubePlaylistService(repo, errHandler, sessionManager)
	// Caso de uso para reordenar playlist
	reorderUseCase := usecases.NewReorderPlaylistUseCase(youtubeService)
	// Handler para operações de playlist
	reorderPlaylist := handlers.NewPlaylistHandler(reorderUseCase, sessionManager)
	getAllPlaylists := handlers.NewGetAllPlaylistsHandler(youtubeService, sessionManager, userRepository)

	// Startanto RabbitMQConsumer
	consumer := messaging.NewRabbitMQConsumer(youtubeService, errHandler)
	go consumer.StartRabbitMQConsumer("reorderApi")

	// Configuração das rotas com Gorilla/mux
	router := routes.ConfigureRoutes(authHandler, reorderPlaylist, getAllPlaylists, sessionManager, authService, userRepository)

	log.Println("API iniciada na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
