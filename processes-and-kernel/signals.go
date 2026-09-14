package main

import (
	"context"
	"fmt"
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
		fmt.Fprintf(os.Stderr, "error finding process: %s\n", err)
		os.Exit(1)
	}

	go func() {
		p.Signal(syscall.SIGHUP)
		p.Signal(syscall.SIGTERM)
		p.Signal(syscall.SIGABRT)

		p.Signal(syscall.SIGINT)
	}()

	<-ctx.Done()

	fmt.Printf("interrupt signal received, exiting...\n")
}
