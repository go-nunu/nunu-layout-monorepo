package server

import (
	"github.com/gin-gonic/gin"
	"nunu-layout-monorepo/app/home/internal/middleware"
	"nunu-layout-monorepo/app/home/internal/router"
	pkghttp "nunu-layout-monorepo/pkg/server/http"
)

func NewHTTPServer(deps router.RouterDeps) *pkghttp.Server {
	if deps.Config.GetString("env") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	s := pkghttp.NewServer(
		gin.New(),
		deps.Logger,
		pkghttp.WithServerHost(deps.Config.GetString("http.host")),
		pkghttp.WithServerPort(deps.Config.GetInt("http.port")),
	)

	s.Use(
		gin.Recovery(),
		middleware.SecurityHeaders(),
		middleware.RequestLog(deps.Logger),
	)

	router.InitSiteRouter(deps, s.Engine)

	return s
}
