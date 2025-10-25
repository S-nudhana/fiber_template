package service

import (
	"context"
	
	"github.com/S-nudhana/fiber_template/internal/core/port"
)

type UserService interface {
	Login(ctx context.Context, email string, password string) (loginStatus bool, uid string, err error)
	Register(ctx context.Context, email string, password string, firstname string, lastname string) (registerStatus bool, err error)
}

type UserServiceImpl struct {
	userRepo port.UserRepository
}

func NewUserService(userRepo port.UserRepository) UserService {
	return &UserServiceImpl{userRepo: userRepo}
}

func (s *UserServiceImpl) Login(ctx context.Context, email string, password string) (loginStatus bool, uid string, err error) {
	authStatus, uid, err := s.userRepo.AuthenticateUser(email, password)
	if err != nil {
		return false, "",err
	}
	return authStatus, uid, nil
}

func (s *UserServiceImpl) Register(ctx context.Context, email string, password string, firstname string, lastname string) (status bool, err error) {
	registerStatus, err := s.userRepo.CreateUser(email, password, firstname, lastname)
	if err != nil {
		return false, err
	}
	return registerStatus, nil
}