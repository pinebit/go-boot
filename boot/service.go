package boot

import "context"

//go:generate mockery --name Service

// Service interface defines the contract for any service managed by go-boot.
// General rules for Start/Stop functions:
// * Implementation must be thread-safe.
// * The context values passed to Start/Stop must not be kept and used ever after.
// * If the context is cancelled, Start/Stop must terminate immediately.
type Service interface {
	// Start() starts the service, returns nil if the service was already started.
	// The service is treated as started if the function returned no error.
	// Also, see the general rules for the Service interface.
	Start(context.Context) error

	// Stop() stops the service, returns nil if the service was already stopped.
	// After calling Stop(), the service is treated as stopped regardless of returned error.
	// Also, see the general rules for the Service interface.
	Stop(context.Context) error
}
