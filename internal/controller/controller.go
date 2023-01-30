package controller

import (
	chatApi "github.com/ITheCorgi/b2b-chat/pkg/api"
)

type controller struct {
	chatApi.UnimplementedChatServer
	chat IChat
}

func New(chatService IChat) controller {
	return controller{chat: chatService}
}
