package boot

import (
	"context"
	"sync"

	"go.uber.org/multierr"
)

type sequentially struct {
	services []Service
	mu       sync.Mutex
}

var _ Service = (*sequentially)(nil)

// Sequentially() creates a Service that delegates invocations to the given services in the sequential order.
// All the general rules defined boot.Service interface are applied.
func Sequentially(services ...Service) Service {
	return &sequentially{
		services: services,
		mu:       sync.Mutex{},
	}
}

// Start(ctx): delegates to all given services in the sequential order. Service at index 0 starts first.
// If a service at index K returned an error, the sequence breaks and this error is returned.
// Note that previously started services 0...K-1 will remain started. In this case,
// the application is expected to proceed with Stop(ctx), where ctx is likely bound to a timeout.
// Closing ctx will break the sequence immediately (see the general rules for boot.Service).
func (s *sequentially) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, service := range s.services {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// Remember that Start() for an already started service returns nil.
		if err := service.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Stop(ctx): delegeates to all given services in the reverse order. Service at index 0 stops last.
// Errors returned from the services are combined with multierr package and do not break the sequence.
// Closing the ctx will break the sequence immediately (see the general rules for boot.Service).
func (s *sequentially) Stop(ctx context.Context) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index := len(s.services) - 1; index >= 0; index-- {
		if ctx.Err() != nil {
			err = multierr.Append(err, ctx.Err())
			break
		}
		// Remember that Stop() for an already stopped service returns nil.
		stopErr := s.services[index].Stop(ctx)
		if stopErr != nil {
			err = multierr.Append(err, stopErr)
		}
	}
	return
}
