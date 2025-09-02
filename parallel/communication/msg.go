package communication

type MessageInterface[T any] interface {
	SetMsg(msg T)
	GetMsg() T
	SetTime(time int)
	GetTime() int
}
