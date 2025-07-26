package dto

// Data transfer objects
type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateUserReq struct {
	ID    int    `json:"ID"`
	Name  string `json:"name"`
	Role  string `json:"role"`
	Email string `json:"email"`
}
