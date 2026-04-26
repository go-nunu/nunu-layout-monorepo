package service_test

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"nunu-layout-monorepo/app/home/internal/service"
	"nunu-layout-monorepo/pkg/log"
)

func TestSiteServiceReturnsConfiguredManifest(t *testing.T) {
	conf := newTestConfig(t)
	logger := log.NewLog(conf)
	siteService := service.NewSiteService(service.NewService(conf, logger))

	health := siteService.Health()
	require.Equal(t, "home", health.App)
	require.Equal(t, "ok", health.Status)

	meta := siteService.Meta()
	require.Equal(t, "test", meta.Stage)
	require.Equal(t, "Home Test", meta.Title)
	require.Equal(t, "app/home/cmd/server", meta.Entry)

	manifest := siteService.Manifest()
	require.Equal(t, "A focused home test shell.", manifest.Headline)
	require.Len(t, manifest.Features, 3)
	require.Equal(t, "/api/v1/manifest", manifest.Features[2].Route)
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
