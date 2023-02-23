package main

import (
	"flag"
	"log"

	"github.com/ITheCorgi/grpc-chat-room/internal/app"
	"github.com/ITheCorgi/grpc-chat-room/internal/config"
	"go.uber.org/zap"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "config.yaml", "--config ./file_name.yaml")
}

func main() {
	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	zapLogger, err := getLogger(cfg)
	if err != nil {
		log.Fatalln("failed to init logger")
	}

	app.Run(cfg, zapLogger)
}

func getLogger(cfg *config.Config) (zapLogger *zap.Logger, err error) {
	defer func() {
		if zapLogger != nil {
			zapLogger.Sync()
			zapLogger.Info("zap synced")
		}
	}()

	switch cfg.App.Environment {
	case "prod":
		zapLogger, err = zap.NewProduction()
		if err != nil {
			return
		}

		zapLogger.Info("production zap logger started")
	default:
		zapLogger, err = zap.NewDevelopment()
		if err != nil {
			return
		}

		zapLogger.Info("dev zap logger started")
	}

	return
}
