package models

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
	ID        int    `json:"ID"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	IsAdmin   bool   `json:"is_admin"`
	IsRefresh bool   `json:"is_refresh"`
	jwt.RegisteredClaims
}
