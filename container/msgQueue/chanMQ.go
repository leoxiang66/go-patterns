package msgqueue

import (
	"context"
	"fmt"
	"sync"
)

type ChanMQ struct {
	container chan []byte
	live      bool
	mutex     sync.Mutex
	deviceId  string
	capacity  int
}

func NewChanMQ(capacity int, deviceId string) *ChanMQ {
	if capacity < 0 {
		capacity = 0
	}

	return &ChanMQ{
		container: make(chan []byte, capacity),
		live:      true,
		mutex:     sync.Mutex{},
		deviceId:  deviceId,
		capacity:  capacity,
	}
}

func (q *ChanMQ) Enq(msg []byte) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if !q.live {
		return fmt.Errorf("insert Msg to a dead MQ")
	}

	select {
	case q.container <- msg:
		// cl.Debug(fmt.Sprintf("Successfully enqueued message for device ID: %s", q.deviceId))
		return nil
	default:
		// cl.Error(fmt.Sprintf("MQ is full for device ID: %s", q.deviceId))
		return fmt.Errorf("MQ is full")
	}
}

func (q *ChanMQ) Deq(ctx context.Context) ([]byte, error) {
	q.mutex.Lock()
	ch := q.container
	q.mutex.Unlock()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context done. Aborting deq")
	case ret, ok := <-ch:
		if !ok {
			return nil, fmt.Errorf("deq a closed MQ")
		}
		return ret, nil
	}
}

func (q *ChanMQ) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.container)
}

func (q *ChanMQ) Clear() error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for {
		select {
		case <-q.container:
		default:
			return nil
		}
	}
}

func (q *ChanMQ) IsLive() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.live
}

func (q *ChanMQ) Renew() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if !q.live {
		q.live = true
		q.container = make(chan []byte, q.capacity)
		// cl.Debug(fmt.Sprintf("Renewed message queue for device ID: %s", q.deviceId))
	}
}

func (q *ChanMQ) Destroy() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.live {
		close(q.container)
		q.live = false
		// cl.Debug(fmt.Sprintf("Destroyed message queue for device ID: %s", q.deviceId))
	}
}
