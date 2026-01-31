package domain

type UserRegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	Firstname string `json:"firstName" validate:"required,min=1"`
	Lastname  string `json:"lastName" validate:"required,min=1"`
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserUpdateRequest struct {
	Firstname string `json:"firstName" validate:"min=1"`
	Lastname  string `json:"lastName" validate:"min=1"`
}