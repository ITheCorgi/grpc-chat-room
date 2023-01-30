package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ITheCorgi/b2b-chat/internal/config"
	"github.com/ITheCorgi/b2b-chat/internal/controller"
	"github.com/ITheCorgi/b2b-chat/internal/usecase"
	chatApi "github.com/ITheCorgi/b2b-chat/pkg/api"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Run(cfg *config.Config, log *zap.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	_, cancelFunc := context.WithCancel(context.Background())

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			return status.Errorf(codes.Internal, "panic triggered: %v", p)
		}),
	}

	grpcServer := grpc.NewServer(
		middleware.WithUnaryServerChain(recovery.UnaryServerInterceptor(opts...)),
		middleware.WithStreamServerChain(recovery.StreamServerInterceptor(opts...)),
	)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.App.Port))
	if err != nil {
		log.Fatal("error creating tcp listener", zap.Error(err))
	}

	chat := controller.New(usecase.New(log))

	chatApi.RegisterChatServer(grpcServer, chat)
	go grpcServer.Serve(listener)

	log.Info("http service started", zap.String("host", cfg.App.Host), zap.String("port", cfg.App.Port))

	sig := <-sigChan
	log.Info("start graceful shutdown, caught sig", zap.String("signal", sig.String()))

	grpcServer.GracefulStop()
	cancelFunc()
}
