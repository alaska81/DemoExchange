package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"DemoExchange/internal/app"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app, err := app.New()
	if err != nil {
		log.Panic(err)
	}

	if err := app.Run(ctx, cancel); err != nil {
		log.Panic(err)
	}

	<-ctx.Done()

	app.Stop()

	time.Sleep(time.Second)
}
