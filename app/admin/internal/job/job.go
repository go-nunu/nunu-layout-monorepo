package job

import (
	"nunu-layout-monorepo/app/admin/internal/repository"
	"nunu-layout-monorepo/pkg/jwt"
	"nunu-layout-monorepo/pkg/log"
	"nunu-layout-monorepo/pkg/sid"
)

type Job struct {
	logger *log.Logger
	sid    *sid.Sid
	jwt    *jwt.JWT
	tm     repository.Transaction
}

func NewJob(
	tm repository.Transaction,
	logger *log.Logger,
	sid *sid.Sid,
) *Job {
	return &Job{
		logger: logger,
		sid:    sid,
		tm:     tm,
	}
}
