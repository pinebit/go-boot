package boot_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/pinebit/go-boot/boot"
	"github.com/pinebit/go-boot/boot/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSimultaneously_StartStop(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)

	s1 := mocks.NewService(t)
	s2 := mocks.NewService(t)
	s3 := mocks.NewService(t)
	s1.On("Start", mock.Anything).Return(nil)
	s1.On("Stop", mock.Anything).Return(nil)
	s2.On("Start", mock.Anything).Return(nil)
	s2.On("Stop", mock.Anything).Return(nil)
	s3.On("Start", mock.Anything).Return(nil)
	s3.On("Stop", mock.Anything).Return(nil)
	services := boot.Simultaneously(s1, s2, s3)

	err := services.Start(ctx)
	assert.NoError(t, err)

	err = services.Stop(ctx)
	assert.NoError(t, err)
}

func TestSimultaneously_StartError(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	broken := errors.New("broken")

	var wg sync.WaitGroup
	wg.Add(2)

	s1 := mocks.NewService(t)
	s1.On("Start", mock.Anything).Run(func(args mock.Arguments) {
		wg.Done()
		time.Sleep(time.Second)
	}).Return(nil)
	s2 := mocks.NewService(t)
	s2.On("Start", mock.Anything).Run(func(args mock.Arguments) {
		wg.Wait()
	}).Return(broken)
	s3 := mocks.NewService(t)
	s3.On("Start", mock.Anything).Run(func(args mock.Arguments) {
		wg.Done()
		time.Sleep(time.Second)
	}).Return(nil)
	services := boot.Simultaneously(s1, s2, s3)

	err := services.Start(ctx)
	assert.ErrorIs(t, err, broken)
}

func TestSimultaneously_StopErrors(t *testing.T) {
	t.Parallel()

	ctx := testContext(t)
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	s1 := mocks.NewService(t)
	s1.On("Stop", mock.Anything).Return(err1)
	s2 := mocks.NewService(t)
	s2.On("Stop", mock.Anything).Return(err2)
	services := boot.Simultaneously(s1, s2)

	err := services.Stop(ctx)
	assert.ErrorContains(t, err, err1.Error())
	assert.ErrorContains(t, err, err2.Error())
}

func TestSimultaneously_StartCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(testContext(t))
	s1 := mocks.NewService(t)
	s1.On("Start", mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(100 * time.Millisecond)
		cancel()
		time.Sleep(100 * time.Millisecond)
	}).Return(nil)
	s2 := mocks.NewService(t)
	s2.On("Start", mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(time.Second)
	}).Return(nil)
	s3 := mocks.NewService(t)
	s3.On("Start", mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(time.Second)
	}).Return(nil)

	services := boot.Simultaneously(s1, s2, s3)
	err := services.Start(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestSimultaneously_StopCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(testContext(t))
	s1 := mocks.NewService(t)
	s1.On("Stop", mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(100 * time.Millisecond)
		cancel()
		time.Sleep(100 * time.Millisecond)
	}).Return(nil)
	s2 := mocks.NewService(t)
	s2.On("Stop", mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(time.Second)
	}).Return(nil)
	s3 := mocks.NewService(t)
	s3.On("Stop", mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(time.Second)
	}).Return(nil)

	services := boot.Simultaneously(s1, s2, s3)
	err := services.Stop(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}
