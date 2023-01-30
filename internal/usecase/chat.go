package usecase

import (
	"context"
	"errors"
	"sync"

	"github.com/ITheCorgi/b2b-chat/internal/entity"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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
		// channel_name <-> channel
		channels map[string]*entity.Chatroom
		// user_name <-> stream message
		connPipe     map[string]chan entity.Message
		withSafeFunc func(mu *sync.RWMutex, isForWrite entity.Lock, fn func() error) error
	}
)

func New(log *zap.Logger) *chat {
	return &chat{
		log: log,

		mu:       &sync.RWMutex{},
		channels: make(map[string]*entity.Chatroom),
		connPipe: make(map[string]chan entity.Message),
		withSafeFunc: func(mu *sync.RWMutex, isForWrite entity.Lock, fn func() error) error {
			if isForWrite == entity.SafeWrite {
				mu.Lock()
				defer mu.Unlock()
			} else {
				mu.RLock()
				defer mu.RUnlock()
			}

			if err := fn(); err != nil {
				return err
			}

			return nil
		},
	}
}

func (c *chat) Connect(ctx context.Context, userName string) error {
	queue := make(chan entity.Message, 100)

	for {
		select {
		case <-ctx.Done():
			c.log.Info("connection closed, disconnecting", zap.String("user_name", userName))

			return nil

		default:
			if err := c.withSafeFunc(c.mu, entity.SafeWrite, func() error {
				c.connPipe[userName] = queue

				//chatroom, err := c.createAndSubscribe(userName, userName, entity.OneToOne)
				//if err != nil {
				//	return err
				//}
				//
				//c.addChatRoom(chatroom)

				return nil
			}); err != nil {
				c.log.Error("failed to create user chat", zap.Error(err))
				return err
			}

		}
	}
}

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

func (c *chat) ListChannels(ctx context.Context) (entity.Channels, error) {
	res := make(entity.Channels, len(c.channels))

	c.withSafeFunc(c.mu, entity.SafeRead, func() error {
		for k := range c.channels {
			res = append(res, entity.Channel{
				Name: c.channels[k].Name,
				Type: c.channels[k].Type,
			})
		}

		return nil
	})

	return res, ctx.Err()
}

func (c *chat) SendMessage(ctx context.Context, message entity.Message, userName string) error {
	if err := c.withSafeFunc(c.mu, entity.SafeWrite, func() error {
		var (
			chatroom                 *entity.Chatroom
			ch                       chan entity.Message
			sentToRoom, sentToFriend bool
		)

		eg, _ := errgroup.WithContext(ctx)

		eg.Go(func() error {
			chatroom = c.isChatExist(message.To)
			if chatroom == nil {
				return nil
			}

			isBelongs := chatroom.IsSubscribed(userName)
			if !isBelongs {
				return nil
			}

			c.distributeMessage(message, chatroom.GetSubscribers())
			sentToRoom = true
			return nil
		})
		eg.Go(func() error {
			ch = c.isUserConnected(message.To)
			if ch == nil {
				return nil
			}

			c.distributeMessage(message, []string{message.To, userName})
			sentToFriend = true
			return nil
		})
		if err := eg.Wait(); err != nil {
			return err
		}

		notFound := !sentToRoom && !sentToFriend
		if notFound {
			return errDestinationAddrDoesntExist
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

}