package sessions

import (
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
	"net/http"
	"time"
)

type SessionManager interface {
	Get(r *http.Request) (map[interface{}]interface{}, error)
	Save(r *http.Request, w http.ResponseWriter, values map[interface{}]interface{}) error
	Remove(r *http.Request, w http.ResponseWriter, key string) error
	GetValue(r *http.Request, key string) (interface{}, error)
	ConversionToString(value interface{}) string
	ConversionToTime(value interface{}) time.Time
	GetUserId(r *http.Request) string
	DestroySession(r *http.Request, w http.ResponseWriter) error
}

type gorillaSessionManager struct {
	store       *sessions.CookieStore
	sessionName string
	secretKey   string
}

func NewGorillaSessionManager(sessionName string, secretKey string) SessionManager {
	store := sessions.NewCookieStore([]byte(secretKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	gothic.Store = store

	return &gorillaSessionManager{
		store:       store,
		sessionName: sessionName,
		secretKey:   secretKey,
	}
}

func (s *gorillaSessionManager) Get(r *http.Request) (map[interface{}]interface{}, error) {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		return nil, err
	}

	return session.Values, nil
}

func (s *gorillaSessionManager) Save(r *http.Request, w http.ResponseWriter, values map[interface{}]interface{}) error {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		return err
	}

	for k, v := range values {
		session.Values[k] = v
	}

	return s.store.Save(r, w, session)
}

func (s *gorillaSessionManager) Remove(r *http.Request, w http.ResponseWriter, key string) error {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		return err
	}

	delete(session.Values, key)
	return session.Save(r, w)
}

func (s *gorillaSessionManager) DestroySession(r *http.Request, w http.ResponseWriter) error {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		return err
	}

	session.Options.MaxAge = -1

	return s.store.Save(r, w, session)
}

func (s *gorillaSessionManager) GetValue(r *http.Request, key string) (interface{}, error) {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		return nil, err
	}

	return session.Values[key], nil
}

func (s *gorillaSessionManager) ConversionToString(value interface{}) string {
	if value == nil {
		return ""
	}

	return value.(string)
}

func (s *gorillaSessionManager) ConversionToTime(value interface{}) time.Time {
	if value == nil {
		return time.Time{}
	}

	str := s.ConversionToString(value)

	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return time.Time{}
	}

	return t
}

func (s *gorillaSessionManager) GetUserId(r *http.Request) string {
	values, err := s.Get(r)
	if err != nil {
		return ""
	}

	return s.ConversionToString(values["user_id"])
}
