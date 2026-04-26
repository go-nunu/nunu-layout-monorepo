//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/spf13/viper"
	"nunu-layout-monorepo/app/admin/internal/repository"
	"nunu-layout-monorepo/app/admin/internal/server"
	"nunu-layout-monorepo/pkg/app"
	"nunu-layout-monorepo/pkg/log"
	"nunu-layout-monorepo/pkg/sid"
)

var repositorySet = wire.NewSet(
	repository.NewDB,
	//repository.NewRedis,
	repository.NewRepository,
	repository.NewCasbinEnforcer,
)
var serverSet = wire.NewSet(
	server.NewMigrateServer,
)

// build App
func newApp(
	migrateServer *server.MigrateServer,
) *app.App {
	return app.NewApp(
		app.WithServer(migrateServer),
		app.WithName("demo-migrate"),
	)
}

func NewWire(*viper.Viper, *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		repositorySet,
		serverSet,
		sid.NewSid,
		newApp,
	))
}
