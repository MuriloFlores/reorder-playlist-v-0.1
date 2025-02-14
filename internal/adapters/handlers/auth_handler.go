package handlers

import (
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"go.uber.org/zap"
	"html/template"
	"log"
	"net/http"
	"project/internal/core/entities"
	"project/internal/core/services"
	"project/internal/infrastructure/logging"
	"project/internal/infrastructure/repository"
	"project/internal/infrastructure/sessions"
)

type handler struct {
	auth  services.AuthService
	store sessions.SessionManager
	repo  repository.UserRepositoryInterface
}

type AuthHandler interface {
	OAuthLogin(http.ResponseWriter, *http.Request)
	OAuthCallback(http.ResponseWriter, *http.Request)
	OAuthLogout(w http.ResponseWriter, r *http.Request)
}

func NewHandler(auth services.AuthService, store sessions.SessionManager, repo repository.UserRepositoryInterface) AuthHandler {
	return &handler{auth: auth, store: store, repo: repo}
}

func (h *handler) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	h.auth.BeginAuthHandler(w, r)
	logging.Info("OAuth Login Handler", zap.String("BeginAuth complete", ""))
}

func (h *handler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		logging.Error("OAuthCallback - auth_handler", zap.String("authentication error: ", err.Error()))
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	var credentials = make(map[interface{}]interface{})
	credentials["user_id"] = user.UserID

	err = h.store.Save(r, w, credentials)
	if err != nil {
		logging.Error("OAuthCallback - save", zap.String("save in session error: ", err.Error()))
		return
	}

	foundUser, err := h.repo.GetUserByID(user.UserID)
	if err != nil {
		logging.Error("OAuthCallback - repo.GetUserByID", zap.String("Find error: ", err.Error()))
		return
	}

	if foundUser == nil {
		logging.Info("OAuthCallback - user not found", zap.String("Creating new user: ", user.UserID))
		err := h.repo.CreateUser(entities.NewUser(
			user.UserID,
			user.Name,
			user.Email,
			user.AccessToken,
			user.RefreshToken,
			user.ExpiresAt,
		))

		if err != nil {
			logging.Error("OAuthCallback - repo.CreateUser", zap.String("Find error: ", err.Error()))
			return
		}
	}

	h.htmlTemplateResponse(w, r, user)

	logging.Info("OAuthCallback - auth_handler", zap.String("complete", user.Name))
}

func (h *handler) OAuthLogout(w http.ResponseWriter, r *http.Request) {
	logging.Info("OAuth Logout", zap.String("Init Logout process", ""))
	err := h.auth.LogoutHandler(w, r)
	if err != nil {
		return
	}

	err = h.store.DestroySession(r, w)
	if err != nil {
		logging.Info("OAuthLogout - auth_handler", zap.String("destroy session error", err.Error()))
		return
	}

	logging.Info("OAuthLogout - auth_handler", zap.String("logout process complete", ""))

	http.Redirect(w, r, "http://localhost:5173/", http.StatusPermanentRedirect)
}

func (h *handler) htmlTemplateResponse(w http.ResponseWriter, r *http.Request, user goth.User) {
	type AuthData struct {
		Token      string
		UserID     string
		UserName   string
		UserEmail  string
		UserAvatar string
	}

	data := AuthData{
		Token:      user.AccessToken,
		UserID:     user.UserID,
		UserName:   user.Name,
		UserEmail:  user.Email,
		UserAvatar: user.AvatarURL,
	}

	const tmpl = `
<!DOCTYPE html>
<html lang="pt-BR">
<head>
<meta charset="UTF-8">
<title>Autenticação Concluída</title>
<script type="text/javascript">
	window.onload = function() {
		var token = "{{.Token}}";
		var user = {
		id: "{{.UserID}}",
		name: "{{.UserName}}",
		email: "{{.UserEmail}}",
		avatar: "{{.UserAvatar}}"
	};
	if (window.opener) {
		window.opener.postMessage({ token: token, user: user }, "http://localhost:5173");
	}

	document.getElementById("message").innerText = "Autenticação concluída. Fechando...";
	setTimeout(function() {
		window.close();
	}, 500);
	}; 
</script>
</head>	
<body> 
	<h1>Autenticação Concluída</h1>
	<p id="message">Processando...</p>
</body>	
</html>`

	t, err := template.New("callback").Parse(tmpl)
	if err != nil {
		http.Error(w, "Erro interno ao processar template", http.StatusInternalServerError)
		log.Printf("Erro ao parsear template: %v", err)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Erro interno ao renderizar template", http.StatusInternalServerError)
		log.Printf("Erro ao executar template: %v", err)
		return
	}

}
