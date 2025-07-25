package models

import "time"

type User struct {
	ID         int       `json:"ID"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	password   string    `json:"-"`
	Created_At time.Time `json:"created_at"`
	Updated_At time.Time `json:"updated_at,omitempty"`
	IsAdmin    bool      `json:"is_admin"`
}

func (u *User) GetPassword() string {
	return u.password
}

func (u *User) SetPassword(password string) {
	u.password = password
}
