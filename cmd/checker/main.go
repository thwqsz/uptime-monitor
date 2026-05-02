package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/thwqsz/uptime-monitor/internal/checker"
	"github.com/thwqsz/uptime-monitor/internal/grpc/checkerpb"
	"google.golang.org/grpc"
)

func main() {
	httpChecker := checker.NewHTTPChecker(&http.Client{})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		httpChecker.RunConsumerLoop(ctx)
		close(done)
	}()

	lis, err := net.Listen("tcp", ":8088")
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	grpcChecker := checker.NewGRPCServerChecker(httpChecker)
	checkerpb.RegisterCheckerServiceServer(srv, grpcChecker)
	go func() {
		log.Println("server started on port :8088")
		if err = srv.Serve(lis); err != nil {
			log.Println(err)
		}
	}()
	chOS := make(chan os.Signal, 1)
	signal.Notify(chOS, syscall.SIGINT, syscall.SIGTERM)
	<-chOS
	log.Println("got stop signal")
	cancel()
	srv.GracefulStop()
	<-done
}
