package controller

import (
	"context"

	"github.com/ITheCorgi/b2b-chat/internal/usecase"
	chatApi "github.com/ITheCorgi/b2b-chat/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type controller struct {
	chatApi.UnimplementedChatServer
	chat usecase.IChat
}

func New(chatService usecase.IChat) controller {
	return controller{chat: chatService}
}

func (c controller) Connect(req *chatApi.ConnectRequest, stream chatApi.Chat_ConnectServer) error {
	if err := req.Validate(); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	return nil
}

func (c controller) CreateGroupChat(ctx context.Context, req *chatApi.GroupChannelNameRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userName, err := getAuthorizationFromMD(ctx)
	if err != nil {
		return nil, err
	}

	if err = c.chat.CreateGroupChat(ctx, req.GetGroupChannelName(), userName); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (c controller) JoinGroupChat(ctx context.Context, req *chatApi.GroupChannelNameRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userName, err := getAuthorizationFromMD(ctx)
	if err != nil {
		return nil, err
	}
}

func (c controller) LeaveGroupChat(ctx context.Context, req *chatApi.GroupChannelNameRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userName, err := getAuthorizationFromMD(ctx)
	if err != nil {
		return nil, err
	}
}

func (c controller) ListChannels(ctx context.Context, _ *emptypb.Empty) (*chatApi.Channels, error) {

}

func (c controller) SendMessage(ctx context.Context, req *chatApi.ChatMessage) (*emptypb.Empty, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userName, err := getAuthorizationFromMD(ctx)
	if err != nil {
		return nil, err
	}
}

func getAuthorizationFromMD(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.DataLoss, "failed to get metadata")
	}

	token := md.Get("authorization")
	if len(token) < 1 {
		return "", status.Errorf(codes.Unauthenticated, "empty authorization field")
	}

	return token[0], nil
}
