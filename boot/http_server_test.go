package boot_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/pinebit/go-boot/boot"
	"github.com/stretchr/testify/assert"
)

func TestHttpServer(t *testing.T) {
	t.Parallel()

	server := &http.Server{
		Addr: ":0",
	}
	serverAsService := boot.NewHttpServer(server)
	app := boot.NewApplicationForService(serverAsService, 5*time.Second)

	ctx, cancel := context.WithCancel(testContext(t))

	var wg sync.WaitGroup
	var finalErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		finalErr = app.Run(ctx)
	}()

	time.Sleep(time.Second)
	cancel()
	wg.Wait()

	assert.NoError(t, finalErr)
}
