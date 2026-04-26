//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/spf13/viper"
	"nunu-layout-monorepo/app/admin/internal/handler"
	"nunu-layout-monorepo/app/admin/internal/job"
	"nunu-layout-monorepo/app/admin/internal/repository"
	"nunu-layout-monorepo/app/admin/internal/server"
	"nunu-layout-monorepo/app/admin/internal/service"
	"nunu-layout-monorepo/pkg/app"
	"nunu-layout-monorepo/pkg/jwt"
	"nunu-layout-monorepo/pkg/log"
	"nunu-layout-monorepo/pkg/server/http"
	"nunu-layout-monorepo/pkg/sid"
)

var repositorySet = wire.NewSet(
	repository.NewDB,
	//repository.NewRedis,
	repository.NewRepository,
	repository.NewTransaction,
	repository.NewUserRepository,
	repository.NewCasbinEnforcer,
	repository.NewAdminRepository,
)

var serviceSet = wire.NewSet(
	service.NewService,
	service.NewUserService,
	service.NewAdminService,
)

var handlerSet = wire.NewSet(
	handler.NewHandler,
	handler.NewUserHandler,
	handler.NewAdminHandler,
)

var jobSet = wire.NewSet(
	job.NewJob,
	job.NewUserJob,
)
var serverSet = wire.NewSet(
	server.NewHTTPServer,
	server.NewJobServer,
)

// build App
func newApp(
	httpServer *http.Server,
	jobServer *server.JobServer,
	// task *server.Task,
) *app.App {
	return app.NewApp(
		app.WithServer(httpServer, jobServer),
		app.WithName("demo-server"),
	)
}

func NewWire(*viper.Viper, *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		repositorySet,
		serviceSet,
		handlerSet,
		jobSet,
		serverSet,
		sid.NewSid,
		jwt.NewJwt,
		newApp,
	))
}
