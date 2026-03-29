package sse

import (
	"sync"

	"github.com/KKKKKKKEM/flowkit/core"
)

type TrackerProvider struct {
	session *Session
}

func (s *TrackerProvider) Track(tag string, meta map[string]any) core.Tracker {
	if meta == nil {
		meta = make(map[string]any)
	}
	meta["tag"] = tag
	return &Tracker{
		meta:    meta,
		session: s.session,
	}
}

// Wait is a no-op for SSE: the SSE event loop handles completion signaling.
func (s *TrackerProvider) Wait() {
}

type Tracker struct {
	mu      sync.Mutex
	session *Session
	meta    map[string]any
}

func (s *Tracker) Update(d map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range d {
		s.meta[k] = v
	}
}

func (s *Tracker) Flush() {
	s.session.Emit(Track, s.meta)

}

func (s *Tracker) Done() {
	s.Update(map[string]any{"status": "done"})
	s.session.Emit(Track, s.meta)
}

func NewSSETrackerProvider(session *Session) *TrackerProvider {
	return &TrackerProvider{session: session}
}
