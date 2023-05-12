package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pinebit/go-boot/boot"
)

// This is an example of using go-boot to build a minimalistic graceful HTTP server.

func main() {
	server := &http.Server{
		Addr: ":8080",
	}
	serverAsService := boot.NewHttpServer(server)
	app := boot.NewApplicationForService(serverAsService, 5*time.Second)
	if err := app.Run(context.Background()); err != nil {
		fmt.Println("server error:", err)
	}
}
