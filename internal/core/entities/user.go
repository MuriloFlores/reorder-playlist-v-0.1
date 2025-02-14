package entities

import "time"

type user struct {
	id           string
	name         string
	email        string
	token        string
	refreshToken string
	expiresAt    time.Time
}

type UserInterface interface {
	Id() string
	Name() string
	Email() string
	Token() string
	RefreshToken() string
	ExpiresAt() time.Time
	SetAccessToken(token string)
	SetExpiresAt(expiresAt time.Time)
	SetRefreshToken(token string)
}

func NewUser(id, name, email, token, refreshToken string, expiresAt time.Time) UserInterface {
	return &user{
		id:           id,
		name:         name,
		email:        email,
		token:        token,
		refreshToken: refreshToken,
		expiresAt:    expiresAt,
	}
}

func (u *user) Id() string {
	return u.id
}

func (u *user) Name() string {
	return u.name
}

func (u *user) Email() string {
	return u.email
}

func (u *user) Token() string {
	return u.token
}

func (u *user) RefreshToken() string {
	return u.refreshToken
}

func (u *user) ExpiresAt() time.Time {
	return u.expiresAt
}

func (u *user) SetExpiresAt(expiresAt time.Time) {
	u.expiresAt = expiresAt
}

func (u *user) SetAccessToken(token string) {
	u.token = token
}

func (u *user) SetRefreshToken(refreshToken string) {
	u.refreshToken = refreshToken
}
