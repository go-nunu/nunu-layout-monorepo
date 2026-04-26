package job

import (
	"context"
	"nunu-layout-monorepo/app/admin/internal/repository"
)

type UserJob interface {
	KafkaConsumer(ctx context.Context) error
}

func NewUserJob(
	job *Job,
	userRepo repository.UserRepository,
) UserJob {
	return &userJob{
		userRepo: userRepo,
		Job:      job,
	}
}

type userJob struct {
	userRepo repository.UserRepository
	*Job
}

func (t userJob) KafkaConsumer(ctx context.Context) error {
	// do something
	return nil
}
