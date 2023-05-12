package boot

import (
	"context"
	"net"
	"net/http"
	"sync"
)

// HttpServer is a Service wrapper for the standard http.Server.
type HttpServer interface {
	Service
}

type httpServer struct {
	server  *http.Server
	started sync.Mutex
}

var _ HttpServer = (*httpServer)(nil)

func NewHttpServer(server *http.Server) HttpServer {
	return &httpServer{
		server: server,
	}
}

func (s *httpServer) Start(ctx context.Context) error {
	if !s.started.TryLock() {
		return nil
	}

	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		s.started.Unlock()
		return err
	}

	go func() {
		_ = s.server.Serve(listener)
	}()

	return nil
}

func (s *httpServer) Stop(ctx context.Context) (err error) {
	defer s.started.Unlock()

	if !s.started.TryLock() {
		err = s.server.Shutdown(ctx)
	}

	return
}
