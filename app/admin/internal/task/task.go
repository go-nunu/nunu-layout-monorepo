package task

import (
	"nunu-layout-monorepo/app/admin/internal/repository"
	"nunu-layout-monorepo/pkg/jwt"
	"nunu-layout-monorepo/pkg/log"
	"nunu-layout-monorepo/pkg/sid"
)

type Task struct {
	logger *log.Logger
	sid    *sid.Sid
	jwt    *jwt.JWT
	tm     repository.Transaction
}

func NewTask(
	tm repository.Transaction,
	logger *log.Logger,
	sid *sid.Sid,
) *Task {
	return &Task{
		logger: logger,
		sid:    sid,
		tm:     tm,
	}
}
