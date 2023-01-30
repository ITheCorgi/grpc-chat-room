package entity

type (
	Channels []Channel

	Channel struct {
		Name string
		Type uint8
	}
)
