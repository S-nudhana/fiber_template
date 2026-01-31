package service

import (
	"context"

	"github.com/S-nudhana/fiber_template/internal/core/port"
)

type UserService interface {
	OAuthLogin(ctx context.Context, email string, provider string, firstName string, lastName string) (loginStatus bool, uid string, err error)
	Login(ctx context.Context, email string, password string) (loginStatus bool, uid string, err error)
	Register(ctx context.Context, email string, password string, firstName string, lastName string) (registerStatus bool, err error)
	DeleteUser(ctx context.Context, uid string) (deleteStatus bool, err error)
	UpdateUser(ctx context.Context, uid string, firstName string, lastName string) (updateStatus bool, err error)
}

type UserServiceImpl struct {
	userRepo port.UserRepository
}

func NewUserService(userRepo port.UserRepository) UserService {
	return &UserServiceImpl{userRepo: userRepo}
}

func (s *UserServiceImpl) OAuthLogin(ctx context.Context, email string, provider string, firstName string, lastName string) (loginStatus bool, uid string, err error) {
	status, uid, err := s.userRepo.OAuthAuthenticateUser(email, provider, firstName, lastName)
	if err != nil {
		return false, "", err
	}
	return status, uid, nil
}

func (s *UserServiceImpl) Login(ctx context.Context, email string, password string) (loginStatus bool, uid string, err error) {
	status, uid, err := s.userRepo.AuthenticateUser(email, password)
	if err != nil {
		return false, "", err
	}
	return status, uid, nil
}

func (s *UserServiceImpl) Register(ctx context.Context, email string, password string, firstName string, lastName string) (registerStatus bool, err error) {
	status, err := s.userRepo.CreateUser(email, password, firstName, lastName)
	if err != nil {
		return false, err
	}
	return status, nil
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, uid string) (deleteStatus bool, err error) {
	status, err := s.userRepo.RemoveUser(uid)
	if err != nil {
		return false, err
	}
	return status, nil
}

func (s *UserServiceImpl) UpdateUser(ctx context.Context, uid string, firstName string, lastName string) (updateStatus bool, err error) {
	status, err := s.userRepo.UpdateUserInfo(uid, firstName, lastName)
	if err != nil {
		return false, err
	}
	return status, nil
}