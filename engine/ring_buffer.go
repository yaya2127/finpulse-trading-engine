package engine

import (
	"errors"
	"sync/atomic"
)

// OrderMessage represents an ultra-low-latency order packet
type OrderMessage struct {
	ID        uint64
	Price     float64
	Quantity  float64
	Side      string // BUY or SELL
	OrderType string // LIMIT, MARKET, IOC, TWAP, VWAP
	Timestamp int64
}

// LockFreeRingBuffer is a lock-free single-producer single-consumer ring buffer
type LockFreeRingBuffer struct {
	buffer []OrderMessage
	capacity uint64
	mask     uint64
	head     uint64
	tail     uint64
}

// NewLockFreeRingBuffer initializes a lock-free ring buffer of power-of-two size
func NewLockFreeRingBuffer(size uint64) *LockFreeRingBuffer {
	if size&(size-1) != 0 {
		size = 1024 // Default to 1024 if not power of 2
	}
	return &LockFreeRingBuffer{
		buffer:   make([]OrderMessage, size),
		capacity: size,
		mask:     size - 1,
	}
}

// Push enqueues an OrderMessage into the ring buffer without locking
func (rb *LockFreeRingBuffer) Push(msg OrderMessage) error {
	head := atomic.LoadUint64(&rb.head)
	tail := atomic.LoadUint64(&rb.tail)

	if tail-head >= rb.capacity {
		return errors.New("ring buffer full")
	}

	rb.buffer[tail&rb.mask] = msg
	atomic.AddUint64(&rb.tail, 1)
	return nil
}

// Pop dequeues an OrderMessage from the ring buffer without locking
func (rb *LockFreeRingBuffer) Pop() (OrderMessage, error) {
	head := atomic.LoadUint64(&rb.head)
	tail := atomic.LoadUint64(&rb.tail)

	if head == tail {
		return OrderMessage{}, errors.New("ring buffer empty")
	}

	msg := rb.buffer[head&rb.mask]
	atomic.AddUint64(&rb.head, 1)
	return msg, nil
}
