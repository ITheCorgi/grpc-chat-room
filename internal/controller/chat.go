package controller

import (
	"context"

	"github.com/ITheCorgi/b2b-chat/internal/entity"
	chatApi "github.com/ITheCorgi/b2b-chat/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (c controller) Connect(req *chatApi.ConnectRequest, stream chatApi.Chat_ConnectServer) error {
	if err := req.Validate(); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	queue, err := c.chat.Connect(stream.Context(), req.GetUsername())
	if err != nil {
		return err
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil

		case msg, _ := <-queue:
			err = stream.Send(convertOutMessage(msg))
			if err != nil {
				return err
			}
		}
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
		return nil, status.Error(codes.Internal, err.Error())
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

	err = c.chat.JoinGroupChat(ctx, req.GetGroupChannelName(), userName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (c controller) LeaveGroupChat(ctx context.Context, req *chatApi.GroupChannelNameRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userName, err := getAuthorizationFromMD(ctx)
	if err != nil {
		return nil, err
	}

	err = c.chat.LeaveGroupChat(ctx, req.GetGroupChannelName(), userName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (c controller) ListChannels(ctx context.Context, _ *emptypb.Empty) (*chatApi.Channels, error) {
	channels, err := c.chat.ListChannels(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	items := make([]*chatApi.Channels_Channel, len(channels))
	for i := range channels {
		i := i
		items[i] = &chatApi.Channels_Channel{
			GroupChannelName: channels[i].Name,
			Type:             chatApi.ChannelType(channels[i].Type),
		}
	}

	return &chatApi.Channels{Items: items}, nil
}

func (c controller) SendMessage(ctx context.Context, req *chatApi.ChatMessage) (*emptypb.Empty, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userName, err := getAuthorizationFromMD(ctx)
	if err != nil {
		return nil, err
	}

	msg, err := convertInMessage(req)
	if err != nil {
		return nil, err
	}

	err = c.chat.SendMessage(ctx, msg, userName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
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

func convertInMessage(req *chatApi.ChatMessage) (entity.Message, error) {
	msg := entity.Message{
		Message: req.GetMessage(),
	}

	switch req.Destination.(type) {
	case *chatApi.ChatMessage_GroupChannelName:
		v := req.Destination.(*chatApi.ChatMessage_GroupChannelName)
		if v == nil {
			return entity.Message{}, status.Error(codes.Internal, "group channel name is empty")
		}

		msg.To = v.GroupChannelName
		msg.ChatType = entity.OneToMany

		return msg, nil

	case *chatApi.ChatMessage_Username:
		v := req.Destination.(*chatApi.ChatMessage_Username)
		if v == nil {
			return entity.Message{}, status.Error(codes.Internal, "user name is empty")
		}

		msg.To = v.Username
		msg.ChatType = entity.OneToOne

		return msg, nil
	}

	return entity.Message{}, status.Error(codes.Internal, "wrong destination")
}

func convertOutMessage(req entity.Message) *chatApi.ChatMessage {
	msg := &chatApi.ChatMessage{
		Message: req.Message,
	}

	switch req.ChatType {
	case entity.OneToMany:
		msg.Destination = &chatApi.ChatMessage_GroupChannelName{
			GroupChannelName: req.To,
		}

		return msg

	case entity.OneToOne:
		msg.Destination = &chatApi.ChatMessage_Username{
			Username: req.To,
		}

		return msg
	}

	return nil
}
