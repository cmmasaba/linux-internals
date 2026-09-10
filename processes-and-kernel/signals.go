package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func unixSignals() {
	// stop when interrupt signal is sent
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	signal.Ignore(syscall.SIGHUP, syscall.SIGTERM, syscall.SIGABRT)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		log.Fatalf("error finding process: %s", err)
	}

	go func() {
		p.Signal(syscall.SIGHUP)
		p.Signal(syscall.SIGABRT)
		p.Signal(syscall.SIGABRT)

		p.Signal(syscall.SIGINT)
	} ()

	<-ctx.Done()

	log.Printf("interrupt signal received, exiting...")
}
