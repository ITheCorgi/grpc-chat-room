package entity

import "sync"

type Chatroom struct {
	Channel
	subscribers sync.Map
}

func (c *Chatroom) AddChannelInfo(info Channel) *Chatroom {
	c.Channel = info

	return c
}

func (c *Chatroom) AddSubscriber(user string) (isSucceed bool) {
	isSucceed = true

	if c.SubscribersLen() == 0 {
		s := sync.Map{}
		s.Store(user, struct{}{})
		c.subscribers = s

		return isSucceed
	}

	_, isExist := c.subscribers.LoadOrStore(user, struct{}{})
	if isExist {
		isSucceed = false

		return isSucceed
	}

	return isSucceed
}

func (c *Chatroom) RemoveSubscriber(user string) bool {
	var isSucceed bool

	if c.SubscribersLen() == 0 {
		return isSucceed
	}

	_, isExist := c.subscribers.LoadAndDelete(user)
	if !isExist {
		return isSucceed
	}

	isSucceed = true
	return isSucceed
}

func (c *Chatroom) IsSubscribed(user string) bool {
	_, isExist := c.subscribers.Load(user)
	return isExist
}

func (c *Chatroom) SubscribersLen() uint {
	var subsAmount uint
	c.subscribers.Range(func(key, value any) bool {
		subsAmount++
		return true
	})

	return subsAmount
}

func (c *Chatroom) GetSubscribers() []string {
	subscribers := []string{}

	c.subscribers.Range(func(key, _ any) bool {
		if _, ok := key.(string); ok {
			subscribers = append(subscribers, key.(string))
			return true
		}
		return false
	})

	return subscribers
}
