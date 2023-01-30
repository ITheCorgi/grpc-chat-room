package usecase

import (
	"context"
	"errors"
	"sync"

	"github.com/ITheCorgi/b2b-chat/internal/entity"
	"go.uber.org/zap"
)

var (
	errDuplicateChannelGroupName   = errors.New("duplicate channel name")
	errChannelGroupDoesntExist     = errors.New("channel group with such name is not found")
	errUserIsAlreadyInGroupChannel = errors.New("user is already inside the group channel")
	errUserNotFound                = errors.New("user was not found in the specified group channel")
	errDestinationAddrDoesntExist  = errors.New("channel group or user is not exist")
)

type (
	chat struct {
		log *zap.Logger

		mu *sync.RWMutex
		// channels keeps a list of active chat rooms (map[chat_name]chat
		channels map[string]*entity.Chatroom
		// connPipe is a pool of client grpc connections (map[user_name]stream queue)
		connPipe map[string]chan entity.Message
		// withSafeFunc provides goroutine safe access to pool and channel list
		withSafeFunc func(mu *sync.RWMutex, safe entity.Lock, fn func() error) error
	}
)

func New(log *zap.Logger) *chat {
	return &chat{
		log: log,

		mu:       &sync.RWMutex{},
		channels: make(map[string]*entity.Chatroom),
		connPipe: make(map[string]chan entity.Message),
		withSafeFunc: func(mu *sync.RWMutex, safe entity.Lock, fn func() error) error {
			switch safe {
			case entity.SafeRead:
				mu.RLock()
				defer mu.RUnlock()
			default:
				mu.Lock()
				defer mu.Unlock()
			}

			if err := fn(); err != nil {
				return err
			}

			return nil
		},
	}
}

// Connect establishes connection with server, returns stream of messages
func (c *chat) Connect(ctx context.Context, userName string) (chan entity.Message, error) {
	queue := make(chan entity.Message, 100)
	if err := c.withSafeFunc(c.mu, entity.SafeWrite, func() error {
		c.connPipe[userName] = queue

		return nil
	}); err != nil {
		c.log.Error("failed to create user chat", zap.Error(err))
		return nil, err
	}

	return queue, nil
}

// CreateGroupChat creates a group chat, in case there is one it returns an error
func (c *chat) CreateGroupChat(ctx context.Context, channelName, userName string) error {
	if err := c.withSafeFunc(c.mu, entity.SafeWrite, func() error {
		if chatroom := c.isChatExist(channelName); chatroom != nil {
			return errDuplicateChannelGroupName
		}

		chatRoom, err := c.createAndSubscribe(channelName, userName, entity.OneToMany)
		if err != nil {
			return err
		}

		c.addChatRoom(chatRoom)

		return nil
	}); err != nil {
		c.log.Error("failed to create group chat", zap.Error(err))
		return err
	}

	return ctx.Err()
}

// JoinGroupChat checks whether chat exists, then subscribes user to chat room
func (c *chat) JoinGroupChat(ctx context.Context, channelName, userName string) error {
	if err := c.withSafeFunc(c.mu, entity.SafeRead, func() error {
		channel := c.isChatExist(channelName)
		if channel == nil {
			return errChannelGroupDoesntExist
		}

		isSucceed := channel.AddSubscriber(userName)
		if !isSucceed {
			return errUserIsAlreadyInGroupChannel
		}

		return nil
	}); err != nil {
		c.log.Error("failed to join group chat", zap.Error(err))
		return err
	}

	return ctx.Err()
}

// LeaveGroupChat checks chat for existing, then unsubscribes user from chat
func (c *chat) LeaveGroupChat(ctx context.Context, channelName, userName string) error {
	if err := c.withSafeFunc(c.mu, entity.SafeWrite, func() error {
		channel := c.isChatExist(channelName)
		if channel == nil {
			return errChannelGroupDoesntExist
		}

		isSucceed := channel.RemoveSubscriber(userName)
		if !isSucceed {
			return errUserNotFound
		}

		if channel.SubscribersLen() == 0 {
			delete(c.channels, channelName)
		}

		return nil
	}); err != nil {
		c.log.Error("failed to leave group chat", zap.Error(err))
		return err
	}

	return ctx.Err()
}

// ListChannels provides a list of existing chat rooms
func (c *chat) ListChannels(ctx context.Context) (entity.Channels, error) {
	res := make(entity.Channels, 0, len(c.channels))

	err := c.withSafeFunc(c.mu, entity.SafeRead, func() error {
		for k := range c.channels {
			res = append(res, entity.Channel{
				Name: c.channels[k].Name,
				Type: c.channels[k].Type,
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, ctx.Err()
}

// SendMessage pushes a message to private or public chats
func (c *chat) SendMessage(ctx context.Context, message entity.Message, userName string) error {
	if err := c.withSafeFunc(c.mu, entity.SafeWrite, func() error {
		switch message.ChatType {
		case entity.OneToOne:
			queue, ok := c.connPipe[message.To]
			if !ok {
				return errUserNotFound
			}

			message.To = userName
			queue <- message

		case entity.OneToMany:
			chatroom := c.isChatExist(message.To)
			if chatroom == nil {
				return nil
			}

			isBelongs := chatroom.IsSubscribed(userName)
			if !isBelongs {
				return errUserNotFound
			}

			c.distributeMessage(message, chatroom.GetSubscribers())
		}

		return nil
	}); err != nil {
		c.log.Error("failed to send message", zap.Error(err))
		return err
	}

	return nil
}

func (c *chat) createAndSubscribe(chat, user string, roomType uint8) (*entity.Chatroom, error) {
	chatRoom := new(entity.Chatroom).
		AddChannelInfo(entity.Channel{
			Name: chat,
			Type: roomType,
		})

	isAdded := chatRoom.AddSubscriber(user)
	if !isAdded {
		return nil, errUserIsAlreadyInGroupChannel
	}

	return chatRoom, nil
}

func (c *chat) addChatRoom(chatroom *entity.Chatroom) {
	c.channels[chatroom.Name] = chatroom
}

func (c *chat) isChatExist(channelName string) *entity.Chatroom {
	channel, ok := c.channels[channelName]
	if ok {
		return channel
	}

	return nil
}

func (c *chat) isUserConnected(user string) chan entity.Message {
	ch, ok := c.connPipe[user]
	if !ok {
		return nil
	}

	return ch
}

func (c *chat) distributeMessage(msg entity.Message, subscribers []string) {
	wg := &sync.WaitGroup{}
	for _, subscriber := range subscribers {
		wg.Add(1)
		go func(subscriber string) {
			ch, ok := c.connPipe[subscriber]
			if !ok {
				c.log.Error("failed to send message", zap.String("subscriber", subscriber))
			}

			ch <- msg
		}(subscriber)
	}
}
