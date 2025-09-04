package msgqueue

import (
	"context"
)

type MessageQueueInterface interface {
	Enq(msg []byte) error
	Deq(ctx context.Context) ([]byte, error)
	Len() int
	Clear() error
	IsLive() bool
	Renew()
	Destroy()
}
