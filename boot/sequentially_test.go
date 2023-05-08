package boot_test

import (
	"context"
	"testing"

	"github.com/pinebit/go-boot/boot"
	"github.com/pinebit/go-boot/boot/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
