package auth

import (
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"golang.org/x/oauth2"
	"net/http"
)

// OAuthConfig contém as configurações necessárias para o OAuth.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	Endpoint     oauth2.Endpoint
}

// InitGoth inicializa os provedores OAuth usando Goth.
func InitGoth(config OAuthConfig) {
	store := sessions.NewCookieStore([]byte(config.ClientSecret))
	store.Options = &sessions.Options{
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	gothic.Store = store

	goth.UseProviders(
		google.New(config.ClientID, config.ClientSecret, config.RedirectURL, config.Scopes...),
	)
}
