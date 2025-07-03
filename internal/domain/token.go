package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	Refresh = "refresh_token"
	Access  = "access_token"
)

type TokenPair struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type CustomClaims struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}
