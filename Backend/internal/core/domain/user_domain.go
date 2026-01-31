package domain

type UserRegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	Firstname string `json:"firstname" validate:"required,min=1"`
	Lastname  string `json:"lastname" validate:"required,min=1"`
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserUpdateRequest struct {
	Firstname string `json:"firstname" validate:"min=1"`
	Lastname  string `json:"lastname" validate:"min=1"`
}