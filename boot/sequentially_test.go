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
	ctx := context.Background()

	var sequence []string
	s1 := createServiceMock(t, ctx, "s1", &sequence)
	s2 := createServiceMock(t, ctx, "s2", &sequence)
	s3 := createServiceMock(t, ctx, "s3", &sequence)
	services := boot.Sequentially(s1, s2, s3)

	err := services.Start(ctx)
	assert.NoError(t, err)
	expectedSequence := []string{"s1.Start()", "s2.Start()", "s3.Start()"}
	assert.Equal(t, expectedSequence, sequence)

	// second attempt to Start() shall do nothing
	err = services.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedSequence, sequence)

	err = services.Stop(ctx)
	assert.NoError(t, err)
	expectedSequence = append(expectedSequence, "s3.Stop()", "s2.Stop()", "s1.Stop()")
	assert.Equal(t, expectedSequence, sequence)

	// second attempt to Stop() shall do nothing
	err = services.Stop(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedSequence, sequence)
}

func TestSequentially_StartError(t *testing.T) {
	ctx := context.Background()
	broken := errors.New("broken")

	s1 := mocks.NewService(t)
	s1.On("Start", ctx).Return(nil)
	s1.On("Stop", ctx).Return(nil)
	s2 := mocks.NewService(t)
	s2.On("Start", ctx).Return(broken)
	s3 := mocks.NewService(t)
	services := boot.Sequentially(s1, s2, s3)

	err := services.Start(ctx)
	assert.ErrorIs(t, err, broken)
}

func TestSequentially_StopError(t *testing.T) {
	ctx := context.Background()
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	s1 := mocks.NewService(t)
	s1.On("Start", ctx).Return(nil)
	s1.On("Stop", ctx).Return(err1)
	s2 := mocks.NewService(t)
	s2.On("Start", ctx).Return(nil)
	s2.On("Stop", ctx).Return(err2)
	services := boot.Sequentially(s1, s2)

	err := services.Start(ctx)
	assert.NoError(t, err)

	err = services.Stop(ctx)
	assert.Equal(t, err.Error(), multierr.Combine(err2, err1).Error())
}

func createServiceMock(t *testing.T, ctx context.Context, name string, sequence *[]string) *mocks.Service {
	s := mocks.NewService(t)
	s.On("Start", ctx).Run(func(args mock.Arguments) {
		*sequence = append(*sequence, name+".Start()")
	}).Return(nil)
	s.On("Stop", ctx).Run(func(args mock.Arguments) {
		*sequence = append(*sequence, name+".Stop()")
	}).Return(nil)
	return s
}
