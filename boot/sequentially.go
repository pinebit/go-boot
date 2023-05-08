package boot

import (
	"context"
	"sync"

	"go.uber.org/multierr"
)

type sequentially struct {
	services []Service
	started  []bool
	mu       sync.Mutex
}

var _ Service = (*sequentially)(nil)

// Creates a Service that internally starts other services in sequential order.
func Sequentially(services ...Service) Service {
	return &sequentially{
		services: services,
		started:  make([]bool, len(services)),
		mu:       sync.Mutex{},
	}
}

func (s *sequentially) Start(ctx context.Context) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, service := range s.services {
		if s.started[index] {
			continue
		}

		err = service.Start(ctx)
		if err != nil {
			if stopErr := s.Stop(ctx); stopErr != nil {
				err = multierr.Append(err, stopErr)
			}
			return
		}

		s.started[index] = true
	}

	return
}

func (s *sequentially) Stop(ctx context.Context) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index := len(s.started) - 1; index >= 0; index-- {
		if s.started[index] {
			stopErr := s.services[index].Stop(ctx)
			if stopErr != nil {
				err = multierr.Append(err, stopErr)
			}
			s.started[index] = false
		}
	}

	return
}
