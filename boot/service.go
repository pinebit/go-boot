package boot

import "context"

//go:generate mockery --name Service

// Service interface defines the contract for any services used by go-boot.
type Service interface {
	// Start() starts the service respecting the given context.
	// The context must not be kept by the service and used ever after.
	// If Start() returns an error, the service state must remain intact.
	// Start() for already started service shall do nothing.
	// Implementation must guard against re-entrant invocations.
	Start(context.Context) error

	// Stop() stops the service respecting the given context.
	// Because the shutdown timing is tight, the method shall finish as soon as possible.
	// Regardless of an error, the service is treated as stoppped.
	// Stop() for already stopped service shall do nothing.
	// Implementation must guard against re-entrant invocations.
	Stop(context.Context) error
}
