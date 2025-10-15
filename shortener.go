package link

import (
	"context"
	"fmt"
	"sync"
)

type Shortener struct {
	mu    sync.RWMutex
	links map[Key]Link
}

func (s *Shortener) Shorten(_ context.Context, link Link) (Key, error) {
	var err error
	s.mu.Lock()
	defer s.mu.Unlock()
	if link.Key, err = Shorten(link); err != nil {
		return "", fmt.Errorf("%w: %w", err, ErrBadRequest)
	}
	if _, ok := s.links[link.Key]; ok {
		return "", fmt.Errorf("saving: %w", ErrConflict)
	}
	if s.links == nil {
		s.links = map[Key]Link{}
	}
	s.links[link.Key] = link
	return link.Key, nil
}

func (s *Shortener) Resolve(_ context.Context, key Key) (Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if key.Empty() {
		return Link{}, fmt.Errorf("validating: empty key: %w", ErrBadRequest)
	}
	if err := key.Validate(); err != nil {
		return Link{}, fmt.Errorf("validating: %w: %w", err, ErrBadRequest)
	}
	link, ok := s.links[key]
	if !ok {
		return Link{}, fmt.Errorf("retrieving: %w", ErrNotFound)
	}
	return link, nil
}
