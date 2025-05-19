package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/frolmr/GophKeeper/internal/server/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server, err := app.NewServerApp()
	if err != nil {
		log.Panicf("Failed to create server app: %v", err)
	}

	if err := server.Run(ctx); err != nil {
		log.Panicf("Server error: %v", err)
	}
}
