package xchannel

import (
	"sync"
)

// Boundless functions as a channel that has an unlimited buffer size.
// The zero value for Boundless is NOT ready for use, must use NewBoundless().
type Boundless struct {
	in       chan interface{}
	out      chan interface{}
	overflow []interface{}

	closed sync.Once
}

// NewBoundless returns a Boundless ready for use.
func NewBoundless(bufferSize int) *Boundless {
	b := &Boundless{
		in:  make(chan interface{}),
		out: make(chan interface{}, bufferSize),
	}
	go b.run()
	return b
}

// Close closes the in-channel. Afterwards you should drain the out-channel in order to release all resources.
func (b *Boundless) Close() {
	b.closed.Do(func() { close(b.in) })
}

// Drain drains the out-channel, discarding all remaining items.
// Should be used in combination with Close(), or else the drain will never finish.
func (b *Boundless) Drain() {
	for range b.out {
	}
}

// In gives a channel that should be sent to by the caller.
func (b *Boundless) In() chan<- interface{} {
	return b.in
}

// Out gives a channel that should be received from by the caller.
func (b *Boundless) Out() <-chan interface{} {
	return b.out
}

func (b *Boundless) run() {
READ_LOOP:
	for {
		if len(b.overflow) == 0 {
			// Overflow queue is empty so incoming items can be pushed directly to the outgoing channel.
			// If outgoing channel is full though, push to overflow.
			select {
			case item, ok := <-b.in:
				if !ok {
					break READ_LOOP
				}
				select {
				case b.out <- item:
					// Optimistically push directly to out.
				default:
					b.overflow = append(b.overflow, item)
				}
			}
		} else {
			nextElement := b.overflow[0]
			// Overflow queue is not empty, so any new items get pushed to the back to preserve order.
			select {
			case item, ok := <-b.in:
				if !ok {
					break READ_LOOP
				}
				b.overflow = append(b.overflow, item)
			case b.out <- nextElement:
				b.overflow = b.overflow[1:]
			}
		}
	}

	// Incoming channel has been closed. Empty overflow queue into the outgoing channel.
	// Note: Outgoing channel should be drained by the user to prevent this from never completing.
	for _, nextElement := range b.overflow {
		b.out <- nextElement
	}

	// Close outgoing channel.
	close(b.out)
}
