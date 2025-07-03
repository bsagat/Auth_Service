package domain

import "time"

// Default app settings
type Config struct {
	Env        string // Application environment: local | dev | prod
	Host       string // Application Host
	Port       string // Application Port number
	Secret     string // Token generate secret key
	Addr       string
	Db         DatabaseConf
	RefreshTTL time.Duration // Refresh token expire time
	AccessTTL  time.Duration // Access token expire time
}

type DatabaseConf struct {
	Name     string
	Password string
	Port     string
	UserName string
}
