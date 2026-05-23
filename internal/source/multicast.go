package source

import (
	"context"
	"sync"

	"github.com/your-org/logstream-tail/internal/logevent"
)

// Multicast fans out events from a single input channel to multiple output
// channels, broadcasting each event to all registered subscribers.
type Multicast struct {
	input  <-chan logevent.Event
	mu     sync.RWMutex
	subs   []chan logevent.Event
	bufSize int
}

// NewMulticast creates a Multicast that reads from input and broadcasts to
// all subscribers. bufSize controls the buffer depth of each subscriber channel.
func NewMulticast(input <-chan logevent.Event, bufSize int) *Multicast {
	if bufSize < 1 {
		bufSize = 64
	}
	return &Multicast{
		input:   input,
		bufSize: bufSize,
	}
}

// Subscribe returns a new channel that will receive every event broadcast
// by the Multicast. The channel is closed when Run returns.
func (m *Multicast) Subscribe() <-chan logevent.Event {
	ch := make(chan logevent.Event, m.bufSize)
	m.mu.Lock()
	m.subs = append(m.subs, ch)
	m.mu.Unlock()
	return ch
}

// Run reads events from the input channel and broadcasts each one to all
// current subscribers. It returns when ctx is cancelled or input is closed.
func (m *Multicast) Run(ctx context.Context) {
	defer m.closeAll()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-m.input:
			if !ok {
				return
			}
			m.broadcast(ctx, ev)
		}
	}
}

func (m *Multicast) broadcast(ctx context.Context, ev logevent.Event) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, ch := range m.subs {
		select {
		case ch <- ev:
		case <-ctx.Done():
			return
		}
	}
}

func (m *Multicast) closeAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ch := range m.subs {
		close(ch)
	}
	m.subs = nil
}
