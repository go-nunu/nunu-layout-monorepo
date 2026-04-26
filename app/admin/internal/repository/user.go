package repository

import (
	"context"
	"nunu-layout-monorepo/model"
)

type UserRepository interface {
	GetUser(ctx context.Context, id int64) (*model.User, error)
}

func NewUserRepository(
	repository *Repository,
) UserRepository {
	return &userRepository{
		Repository: repository,
	}
}

type userRepository struct {
	*Repository
}

func (r *userRepository) GetUser(ctx context.Context, id int64) (*model.User, error) {
	var user model.User

	return &user, nil
}
