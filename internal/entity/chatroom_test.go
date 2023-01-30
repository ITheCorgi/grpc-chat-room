package entity

import "testing"

type chat struct {
	room map[string]*Chatroom
}

func Test_AddChannelInfo(t *testing.T) {
	t.Run("test create channel with info", func(t *testing.T) {
		info := Channel{
			Name: "test_channel",
			Type: OneToOne,
		}

		c := new(chat)
		c.room = map[string]*Chatroom{}

		c.room[info.Name] = &Chatroom{}
		exp := c.room[info.Name].AddChannelInfo(info)
		act := c.room[info.Name]

		if exp.Name != act.Name {
			t.Errorf("expected and actual names are different: exp: %v, act: %v", exp, act)
		}
		if exp.Type != act.Type {
			t.Errorf("expected and actual types are different: exp: %v, act: %v", exp, act)
		}
	})
}

func Test_AddSubscriber(t *testing.T) {
	t.Run("test add subscriber to chatroom", func(t *testing.T) {
		info := Channel{
			Name: "test_channel",
			Type: OneToOne,
		}

		c := new(chat)
		c.room = map[string]*Chatroom{}

		c.room[info.Name] = &Chatroom{}
		exp := c.room[info.Name].AddChannelInfo(info)
		act := c.room[info.Name]

		if exp.Name != act.Name {
			t.Errorf("expected and actual names are different: exp: %v, act: %v", exp, act)
		}
		if exp.Type != act.Type {
			t.Errorf("expected and actual types are different: exp: %v, act: %v", exp, act)
		}

		subscriber := "subscriber1"
		isSucceed := c.room[info.Name].AddSubscriber(subscriber)
		if !isSucceed {
			t.Error("failed to add subscriber")
		}

		if c.room[info.Name].SubscribersLen() != 1 {
			t.Error("subscribers len mismatch")
		}
	})
}

func Test_RemoveSubscriber(t *testing.T) {
	t.Run("test remove subscriber", func(t *testing.T) {
		info := Channel{
			Name: "test_channel",
			Type: OneToOne,
		}

		c := new(chat)
		c.room = map[string]*Chatroom{}

		c.room[info.Name] = &Chatroom{}
		exp := c.room[info.Name].AddChannelInfo(info)
		act := c.room[info.Name]

		if exp.Name != act.Name {
			t.Errorf("expected and actual names are different: exp: %v, act: %v", exp, act)
		}
		if exp.Type != act.Type {
			t.Errorf("expected and actual types are different: exp: %v, act: %v", exp, act)
		}

		subscriber := "subscriber1"
		isSucceed := c.room[info.Name].AddSubscriber(subscriber)
		if !isSucceed {
			t.Error("failed to add subscriber")
		}

		if c.room[info.Name].SubscribersLen() != 1 {
			t.Error("subscribers len mismatch")
		}

		isSucceed = c.room[info.Name].RemoveSubscriber(subscriber)
		if !isSucceed {
			t.Error("failed to remove subscriber")
		}
	})
}

func Test_IsSubscribed(t *testing.T) {
	t.Run("test is subscribed user", func(t *testing.T) {
		info := Channel{
			Name: "test_channel",
			Type: OneToOne,
		}

		c := new(chat)
		c.room = map[string]*Chatroom{}

		c.room[info.Name] = &Chatroom{}
		exp := c.room[info.Name].AddChannelInfo(info)
		act := c.room[info.Name]

		if exp.Name != act.Name {
			t.Errorf("expected and actual names are different: exp: %v, act: %v", exp, act)
		}
		if exp.Type != act.Type {
			t.Errorf("expected and actual types are different: exp: %v, act: %v", exp, act)
		}

		subscriber := "subscriber1"
		isSucceed := c.room[info.Name].AddSubscriber(subscriber)
		if !isSucceed {
			t.Error("failed to add subscriber")
		}

		if c.room[info.Name].SubscribersLen() != 1 {
			t.Error("subscribers len mismatch")
		}

		isIn := c.room[info.Name].IsSubscribed(subscriber)
		if !isIn {
			t.Error("failed to check IsSubscribed")
		}
	})
}

func Test_SubscribersLen(t *testing.T) {
	t.Run("test get subscription amount", func(t *testing.T) {
		info := Channel{
			Name: "test_channel",
			Type: OneToOne,
		}

		c := new(chat)
		c.room = map[string]*Chatroom{}

		c.room[info.Name] = &Chatroom{}
		exp := c.room[info.Name].AddChannelInfo(info)
		act := c.room[info.Name]

		if exp.Name != act.Name {
			t.Errorf("expected and actual names are different: exp: %v, act: %v", exp, act)
		}
		if exp.Type != act.Type {
			t.Errorf("expected and actual types are different: exp: %v, act: %v", exp, act)
		}

		subscriberOne := "subscriber1"
		subscriberTwo := "subscriber2"
		isSucceed := c.room[info.Name].AddSubscriber(subscriberOne)
		if !isSucceed {
			t.Error("failed to add subscriber")
		}
		isSucceed = c.room[info.Name].AddSubscriber(subscriberTwo)
		if !isSucceed {
			t.Error("failed to add subscriber")
		}

		if c.room[info.Name].SubscribersLen() != 2 {
			t.Error("subscribers len mismatch")
		}
	})
}

func Test_GetSubscribers(t *testing.T) {
	t.Run("test get subscription amount", func(t *testing.T) {
		info := Channel{
			Name: "test_channel",
			Type: OneToOne,
		}

		c := new(chat)
		c.room = map[string]*Chatroom{}

		c.room[info.Name] = &Chatroom{}
		exp := c.room[info.Name].AddChannelInfo(info)
		act := c.room[info.Name]

		if exp.Name != act.Name {
			t.Errorf("expected and actual names are different: exp: %v, act: %v", exp, act)
		}
		if exp.Type != act.Type {
			t.Errorf("expected and actual types are different: exp: %v, act: %v", exp, act)
		}

		subscriberOne := "subscriber1"
		subscriberTwo := "subscriber2"
		isSucceed := c.room[info.Name].AddSubscriber(subscriberOne)
		if !isSucceed {
			t.Error("failed to add subscriber")
		}
		isSucceed = c.room[info.Name].AddSubscriber(subscriberTwo)
		if !isSucceed {
			t.Error("failed to add subscriber")
		}

		subsList := c.room[info.Name].GetSubscribers()
		if len(subsList) != 2 {
			t.Error("subscription amount mismatch")
		}

		if subsList[0] != subscriberOne && subsList[1] != subscriberTwo {
			t.Error("got wrong subscribers")
		}
	})
}
