package boot

import (
	"context"
	"sync"

	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
)

type simultaneously struct {
	services []Service
	mu       sync.Mutex
}

var _ Service = (*simultaneously)(nil)

// Simultaneously() creates a Service that delegates invocations to the given services simultaneously.
// All the general rules defined boot.Service interface are applied.
func Simultaneously(services ...Service) Service {
	return &simultaneously{
		services: services,
		mu:       sync.Mutex{},
	}
}

// Start(ctx): simultaneously delegates to all given services.
// If a service returns an error, the sequence breaks and this error is returned.
// Note that other started services will remain started. In this case,
// the application is expected to proceed with Stop(ctx), where ctx is usually bound to a timeout.
// Closing ctx will break the sequence immediately (see the general rules for boot.Service).
func (s *simultaneously) Start(ctx context.Context) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, groupCtx := errgroup.WithContext(ctx)
	for _, service := range s.services {
		service := service
		group.Go(func() error {
			return service.Start(groupCtx)
		})
	}

	doneCh := make(chan struct{})
	go func() {
		err = group.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		return
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop(ctx): simultaneously delegeates to all given services.
// Errors returned from the services are combined with multierr package and do not break the sequence.
// Closing the ctx will break the sequence immediately (see the general rules for boot.Service).
func (s *simultaneously) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	errors := make([]error, len(s.services))
	wg := &sync.WaitGroup{}
	wg.Add(len(s.services))

	for index := len(s.services) - 1; index >= 0; index-- {
		index := index
		go func() {
			defer wg.Done()
			errors[index] = s.services[index].Stop(ctx)
		}()
	}

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		return multierr.Combine(errors...)
	case <-ctx.Done():
		return ctx.Err()
	}
}
