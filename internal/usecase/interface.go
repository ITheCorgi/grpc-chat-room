package usecase

import (
	"context"

	"github.com/ITheCorgi/b2b-chat/internal/entity"
)

type IChat interface {
	Connect(ctx context.Context, userName string) (chan entity.Message, error)
	CreateGroupChat(ctx context.Context, channelName, userName string) error
	JoinGroupChat(ctx context.Context, channelName, userName string) error
	LeaveGroupChat(ctx context.Context, channelName, userName string) error
	ListChannels(ctx context.Context) (entity.Channels, error)
	SendMessage(ctx context.Context, message entity.Message, userName string) error
}
