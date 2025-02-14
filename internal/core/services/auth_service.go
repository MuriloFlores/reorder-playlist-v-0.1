package services

import (
	"context"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"net/http"
	"project/internal/infrastructure/config"
	"project/internal/infrastructure/logging"
	"time"
)

type AuthService interface {
	BeginAuthHandler(w http.ResponseWriter, r *http.Request)
	CompleteUserAuth(w http.ResponseWriter, r *http.Request) (goth.User, error)
	LogoutHandler(w http.ResponseWriter, r *http.Request) error
	RefreshAccessToken(refreshToken string) (string, time.Time, error)
}

type GothAuthService struct{}

func NewAuthService() AuthService {
	return &GothAuthService{}
}

func (g *GothAuthService) BeginAuthHandler(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

func (g *GothAuthService) CompleteUserAuth(w http.ResponseWriter, r *http.Request) (goth.User, error) {
	return gothic.CompleteUserAuth(w, r)
}

func (g *GothAuthService) LogoutHandler(w http.ResponseWriter, r *http.Request) error {
	return gothic.Logout(w, r)
}

func (g *GothAuthService) RefreshAccessToken(refreshToken string) (string, time.Time, error) {
	configure := oauth2.Config{
		ClientID:     config.EnvConfigs.ClientID,
		ClientSecret: config.EnvConfigs.SecretKey,
		Endpoint:     google.Endpoint,
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	newToken, err := configure.TokenSource(context.Background(), token).Token()
	if err != nil {
		logging.Info("refreshAccessToken - auth_service - L54", zap.String("refresh token error", err.Error()))
		return "", time.Time{}, nil
	}

	return newToken.AccessToken, newToken.Expiry, nil
}
