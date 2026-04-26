package service

import (
	"nunu-layout-monorepo/app/admin/internal/repository"
	"nunu-layout-monorepo/pkg/jwt"
	"nunu-layout-monorepo/pkg/log"
	"nunu-layout-monorepo/pkg/sid"
)

type Service struct {
	logger *log.Logger
	sid    *sid.Sid
	jwt    *jwt.JWT
	tm     repository.Transaction
}

func NewService(
	tm repository.Transaction,
	logger *log.Logger,
	sid *sid.Sid,
	jwt *jwt.JWT,
) *Service {
	return &Service{
		logger: logger,
		sid:    sid,
		jwt:    jwt,
		tm:     tm,
	}
}
