package service

import (
	"github.com/spf13/viper"
	"nunu-layout-monorepo/pkg/log"
)

type Service struct {
	config *viper.Viper
	logger *log.Logger
}

func NewService(conf *viper.Viper, logger *log.Logger) *Service {
	return &Service{
		config: conf,
		logger: logger,
	}
}
