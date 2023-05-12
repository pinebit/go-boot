package boot

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/multierr"
)

// Application represents an application that supports graceful boot/shutdown.
type Applicaiton interface {
	// Start starts the application. This is a blocking, not thread-safe call.
	// This method handles SIGTERM/SIGINT signals and automatically shutting down the app.
	// The provided user context can be used to signal application to stop gracefully.
	Run(ctx context.Context) error
}

type application struct {
	service         Service
	shutdownTimeout time.Duration
}

var _ Applicaiton = (*application)(nil)

// NewApplicationForService creates a new Application from the given (uber) Service,
// that is usually constructed with Sequentially or Simultaneously (or combination).
// A commonly recommended shutdownTimeout range is 5-15 seconds.
func NewApplicationForService(service Service, shutdownTimeout time.Duration) Applicaiton {
	return &application{
		service:         service,
		shutdownTimeout: shutdownTimeout,
	}
}

func (a application) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	startErr := a.service.Start(ctx)
	if startErr == nil {
		<-ctx.Done()
	}
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	return multierr.Combine(startErr, a.service.Stop(ctx))
}
