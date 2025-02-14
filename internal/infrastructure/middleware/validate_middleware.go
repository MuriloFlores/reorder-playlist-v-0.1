package middleware

import (
	"go.uber.org/zap"
	"net/http"
	"project/internal/core/services"
	"project/internal/infrastructure/logging"
	"project/internal/infrastructure/repository"
	"project/internal/infrastructure/sessions"
	"time"
)

type authMiddleware struct {
	store sessions.SessionManager
	auth  services.AuthService
	repo  repository.UserRepositoryInterface
}

type AuthMiddlewareInterface interface {
	ValidateTokenHandler(next http.Handler) http.Handler
}

func NewAuthMiddleware(store sessions.SessionManager, auth services.AuthService, repo repository.UserRepositoryInterface) AuthMiddlewareInterface {
	return &authMiddleware{
		store: store,
		auth:  auth,
		repo:  repo,
	}
}

type JSONValidatorPayload struct {
	UserId string `json:"user_id"`
}

func (a *authMiddleware) ValidateTokenHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := a.store.GetUserId(r)
		if userId == "" {
			userId = r.Header.Get("X-User-Id")
			if userId == "" {
				logging.Error("GetAllPlaylists - header missing", zap.String("header", "X-User-Id"))
				http.Error(w, "Unauthorized: User ID header missing", http.StatusUnauthorized)
				return
			}
		}

		user, err := a.repo.GetUserByID(userId)
		if err != nil {
			logging.Error("validate_middleware - ValidateTokenHandler", zap.String("consult error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if user == nil {
			logging.Error("validate_middleware - ValidateTokenHandler", zap.String("userNotFound", err.Error()))
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}

		if user.ExpiresAt().IsZero() || user.ExpiresAt().Before(time.Now()) {
			if user.RefreshToken() == "" {
				logging.Error("ValidateToken - auth_handler", zap.String("refresh token error", "empty refresh token or cast error"))
				http.Error(w, "Refresh Token not found", http.StatusUnauthorized)
				return
			}

			newAccessToken, newExpiry, err := a.auth.RefreshAccessToken(user.RefreshToken())
			if err != nil {
				logging.Error("ValidateToken - auth_handler", zap.String("refresh token error", err.Error()))
				http.Error(w, "Refresh Token not found", http.StatusUnauthorized)
				return
			}

			user.SetAccessToken(newAccessToken)
			user.SetExpiresAt(newExpiry)

			err = a.repo.UpdateUser(user)
			if err != nil {
				logging.Error("ValidateToken - auth_handler", zap.String("update user infos error", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
