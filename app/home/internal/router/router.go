package router

import (
	"github.com/spf13/viper"
	"nunu-layout-monorepo/app/home/internal/handler"
	"nunu-layout-monorepo/pkg/log"
)

type RouterDeps struct {
	Logger      *log.Logger
	Config      *viper.Viper
	SiteHandler *handler.SiteHandler
}
