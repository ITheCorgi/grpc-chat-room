package controller

import (
	"context"

	"github.com/ITheCorgi/grpc-chat-room/internal/entity"
)

type IChat interface {
	// Connect establishes connection with server, returns stream of messages
	Connect(ctx context.Context, userName string) (chan entity.Message, error)
	// CreateGroupChat creates a group chat, in case there is one it returns an error
	CreateGroupChat(ctx context.Context, channelName, userName string) error
	// JoinGroupChat checks whether chat exists, then subscribes user to chat room
	JoinGroupChat(ctx context.Context, channelName, userName string) error
	// LeaveGroupChat checks chat for existing, then unsubscribes user from chat
	LeaveGroupChat(ctx context.Context, channelName, userName string) error
	// ListChannels provides a list of existing chat rooms
	ListChannels(ctx context.Context) (entity.Channels, error)
	// SendMessage pushes a message to private or public chats
	SendMessage(ctx context.Context, message entity.Message, userName string) error
}
