package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/thwqsz/uptime-monitor/internal/checker"
)

func main() {
	httpChecker := checker.NewHTTPChecker(&http.Client{})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		httpChecker.RunConsumerLoop(ctx)
		close(done)
	}()

	chOS := make(chan os.Signal, 1)
	signal.Notify(chOS, syscall.SIGINT, syscall.SIGTERM)
	<-chOS
	log.Println("got stop signal")
	cancel()

	<-done
}
