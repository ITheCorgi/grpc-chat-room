package app

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ITheCorgi/b2b-chat/internal/config"
	"github.com/ITheCorgi/b2b-chat/internal/controller"
	"github.com/ITheCorgi/b2b-chat/internal/usecase"
	chatApi "github.com/ITheCorgi/b2b-chat/pkg/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Run(cfg *config.Config, log *zap.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	_, cancelFunc := context.WithCancel(context.Background())

	grpcServer := grpc.NewServer()

	listener, err := net.Listen("tcp", cfg.App.Port)
	if err != nil {
		log.Fatal("error creating tcp listener", zap.Error(err))
	}

	chat := controller.New(usecase.New(log))

	chatApi.RegisterChatServer(grpcServer, chat)
	grpcServer.Serve(listener)

	log.Info("http service started", zap.String("host", cfg.App.Host), zap.String("port", cfg.App.Port))

	sig := <-sigChan
	log.Info("start graceful shutdown, caught sig", zap.String("signal", sig.String()))

	grpcServer.GracefulStop()
	cancelFunc()
}
