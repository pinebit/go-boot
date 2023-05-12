package boot_test

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"time"

	"github.com/pinebit/go-boot/boot"
	"github.com/pinebit/go-boot/boot/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApplication_Run(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	s1 := mocks.NewService(t)
	s2 := mocks.NewService(t)
	s3 := mocks.NewService(t)
	s1.On("Start", mock.Anything).Run(func(args mock.Arguments) {
		s2.On("Start", mock.Anything).Run(func(args mock.Arguments) {
			s3.On("Start", mock.Anything).Return(nil)
		}).Return(nil)
	}).Return(nil)
	s3.On("Stop", mock.Anything).Run(func(args mock.Arguments) {
		s2.On("Stop", mock.Anything).Run(func(args mock.Arguments) {
			s1.On("Stop", mock.Anything).Return(nil)
		}).Return(nil)
	}).Return(nil)
	services := boot.Sequentially(s1, s2, s3)
	app := boot.NewApplicationForService(services, 5*time.Second)

	t.Run("shutting down due to SIGINT", func(t *testing.T) {
		go func() {
			<-time.After(100 * time.Millisecond)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()
		err := app.Run(ctx)

		assert.NoError(t, err)
	})

	t.Run("shutting down due to SIGTERM", func(t *testing.T) {
		go func() {
			<-time.After(100 * time.Millisecond)
			syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()
		err := app.Run(ctx)

		assert.NoError(t, err)
	})

	t.Run("shutting down due to user context", func(t *testing.T) {
		uctx, cancel := context.WithCancel(ctx)
		go func() {
			<-time.After(100 * time.Millisecond)
			cancel()
		}()
		err := app.Run(uctx)

		assert.NoError(t, err)
	})
}

func TestApplication_StartError(t *testing.T) {
	t.Parallel()

	startErr := errors.New("start")
	ctx := testContext(t)
	s1 := mocks.NewService(t)
	s1.On("Start", mock.Anything).Return(startErr)
	s1.On("Stop", mock.Anything).Return(nil)
	app := boot.NewApplicationForService(s1, 5*time.Second)

	err := app.Run(ctx)

	assert.ErrorIs(t, err, startErr)
}

func TestApplication_StopError(t *testing.T) {
	t.Parallel()

	stopErr := errors.New("stop")
	ctx, cancel := context.WithCancel(testContext(t))
	s1 := mocks.NewService(t)
	s1.On("Start", mock.Anything).Return(nil)
	s1.On("Stop", mock.Anything).Return(stopErr)
	app := boot.NewApplicationForService(s1, 5*time.Second)

	cancel()
	err := app.Run(ctx)

	assert.ErrorIs(t, err, stopErr)
}
