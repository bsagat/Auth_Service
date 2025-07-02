package domain

import "time"

// Default app settings
type Config struct {
	Env      string
	Host     string
	Port     string
	Addr     string
	TokenTTL time.Duration
}
