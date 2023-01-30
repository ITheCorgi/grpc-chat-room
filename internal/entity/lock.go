package entity

type Lock bool

const (
	SafeRead  Lock = false
	SafeWrite Lock = true
)
