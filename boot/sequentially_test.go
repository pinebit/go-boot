package boot_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pinebit/go-boot/boot"
	"github.com/pinebit/go-boot/boot/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
)

func TestSequentially_StartStop(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)

	s1 := mocks.NewService(t)
	s2 := mocks.NewService(t)
	s3 := mocks.NewService(t)
	s1.On("Start", ctx).Run(func(args mock.Arguments) {
		s2.On("Start", ctx).Run(func(args mock.Arguments) {
			s3.On("Start", ctx).Return(nil)
		}).Return(nil)
	}).Return(nil)
	s3.On("Stop", ctx).Run(func(args mock.Arguments) {
		s2.On("Stop", ctx).Run(func(args mock.Arguments) {
			s1.On("Stop", ctx).Return(nil)
		}).Return(nil)
	}).Return(nil)
	services := boot.Sequentially(s1, s2, s3)

	err := services.Start(ctx)
	assert.NoError(t, err)

	err = services.Stop(ctx)
	assert.NoError(t, err)
}

func TestSequentially_StartError(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	broken := errors.New("broken")

	s1 := mocks.NewService(t)
	s1.On("Start", ctx).Return(nil)
	s2 := mocks.NewService(t)
	s2.On("Start", ctx).Return(broken)
	s3 := mocks.NewService(t)
	services := boot.Sequentially(s1, s2, s3)

	err := services.Start(ctx)
	assert.ErrorIs(t, err, broken)
}

func TestSequentially_StopErrors(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	s1 := mocks.NewService(t)
	s1.On("Stop", ctx).Return(err1)
	s2 := mocks.NewService(t)
	s2.On("Stop", ctx).Return(err2)
	services := boot.Sequentially(s1, s2)

	err := services.Stop(ctx)
	assert.Equal(t, err.Error(), multierr.Combine(err2, err1).Error())
}

func TestSequentially_StartCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(testContext(t))
	s1 := mocks.NewService(t)
	s1.On("Start", ctx).Run(func(args mock.Arguments) {
		cancel()
	}).Return(nil)
	s2 := mocks.NewService(t)
	s3 := mocks.NewService(t)

	services := boot.Sequentially(s1, s2, s3)
	err := services.Start(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestSequentially_StopCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(testContext(t))
	s1 := mocks.NewService(t)
	s2 := mocks.NewService(t)
	s3 := mocks.NewService(t)
	s3.On("Stop", ctx).Run(func(args mock.Arguments) {
		cancel()
	}).Return(nil)

	services := boot.Sequentially(s1, s2, s3)
	err := services.Stop(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}
