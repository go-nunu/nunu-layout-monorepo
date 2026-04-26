package server_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"nunu-layout-monorepo/app/home/internal/handler"
	"nunu-layout-monorepo/app/home/internal/router"
	homeserver "nunu-layout-monorepo/app/home/internal/server"
	"nunu-layout-monorepo/app/home/internal/service"
	"nunu-layout-monorepo/pkg/log"
)

func TestHomeHTTPServerRoutes(t *testing.T) {
	server := newTestServer(t)

	t.Run("index page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "Public-facing shell")
	})

	t.Run("health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "\"status\":\"ok\"")
		require.Contains(t, resp.Body.String(), "\"app\":\"home\"")
	})

	t.Run("meta", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/meta", nil)
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "\"entry\":\"app/home/cmd/server\"")
		require.Contains(t, resp.Body.String(), "\"title\":\"Home Test\"")
	})

	t.Run("manifest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/manifest", nil)
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "\"headline\":\"A focused home test shell.\"")
		require.Contains(t, resp.Body.String(), "Bootstrap Manifest")
	})
}

func newTestServer(t *testing.T) http.Handler {
	t.Helper()

	conf := newTestConfig(t)
	logger := log.NewLog(conf)
	siteService := service.NewSiteService(service.NewService(conf, logger))
	siteHandler := handler.NewSiteHandler(handler.NewHandler(logger), siteService)
	deps := router.RouterDeps{
		Logger:      logger,
		Config:      conf,
		SiteHandler: siteHandler,
	}

	return homeserver.NewHTTPServer(deps)
}

func newTestConfig(t *testing.T) *viper.Viper {
	t.Helper()

	conf := viper.New()
	conf.Set("env", "test")
	conf.Set("http.host", "127.0.0.1")
	conf.Set("http.port", 8081)
	conf.Set("site.title", "Home Test")
	conf.Set("site.headline", "A focused home test shell.")
	conf.Set("log.log_level", "debug")
	conf.Set("log.mode", "console")
	conf.Set("log.encoding", "console")
	conf.Set("log.log_file_name", filepath.Join(t.TempDir(), "home-test.log"))
	conf.Set("log.max_backups", 1)
	conf.Set("log.max_age", 1)
	conf.Set("log.max_size", 1)
	conf.Set("log.compress", false)

	return conf
}
