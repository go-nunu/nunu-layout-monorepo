//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/spf13/viper"
	"nunu-layout-monorepo/app/home/internal/handler"
	"nunu-layout-monorepo/app/home/internal/router"
	"nunu-layout-monorepo/app/home/internal/server"
	"nunu-layout-monorepo/app/home/internal/service"
	"nunu-layout-monorepo/pkg/app"
	"nunu-layout-monorepo/pkg/log"
	pkghttp "nunu-layout-monorepo/pkg/server/http"
)

var serviceSet = wire.NewSet(
	service.NewService,
	service.NewSiteService,
)

var handlerSet = wire.NewSet(
	handler.NewHandler,
	handler.NewSiteHandler,
)

var serverSet = wire.NewSet(
	server.NewHTTPServer,
)

func newApp(httpServer *pkghttp.Server) *app.App {
	return app.NewApp(
		app.WithServer(httpServer),
		app.WithName("home-server"),
	)
}

func NewWire(*viper.Viper, *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		serviceSet,
		handlerSet,
		serverSet,
		wire.Struct(new(router.RouterDeps), "*"),
		newApp,
	))
}
