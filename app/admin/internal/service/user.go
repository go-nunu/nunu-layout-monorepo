package service

import (
	"context"
	"nunu-layout-monorepo/app/admin/internal/repository"
	"nunu-layout-monorepo/model"
)

type UserService interface {
	GetUser(ctx context.Context, id int64) (*model.User, error)
}

func NewUserService(
	service *Service,
	userRepository repository.UserRepository,
) UserService {
	return &userService{
		Service:        service,
		userRepository: userRepository,
	}
}

type userService struct {
	*Service
	userRepository repository.UserRepository
}

func (s *userService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return s.userRepository.GetUser(ctx, id)
}
