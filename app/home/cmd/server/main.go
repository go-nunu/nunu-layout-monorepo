package main

import (
	"context"
	"flag"
	"fmt"

	"go.uber.org/zap"
	"nunu-layout-monorepo/app/home/cmd/server/wire"
	"nunu-layout-monorepo/pkg/config"
	"nunu-layout-monorepo/pkg/log"
)

func main() {
	var envConf = flag.String("conf", "config/home/local.yml", "config path, eg: -conf ./config/home/local.yml")
	flag.Parse()

	conf := config.NewConfig(*envConf)
	logger := log.NewLog(conf)

	application, cleanup, err := wire.NewWire(conf, logger)
	defer cleanup()
	if err != nil {
		panic(err)
	}

	logger.Info("home server start", zap.String("host", fmt.Sprintf("http://%s:%d", conf.GetString("http.host"), conf.GetInt("http.port"))))
	if err := application.Run(context.Background()); err != nil {
		panic(err)
	}
}
