package main

import (
	"context"
	"github.com/pershin-daniil/internship_backend_2022/internal/logger"
	"github.com/pershin-daniil/internship_backend_2022/internal/server"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	address = ":6321"
	version = "0.0.1"
)

func main() {
	log := logger.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := server.New(log, address, version)

	go func() {
		signCh := make(chan os.Signal, 1)
		signal.Notify(signCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
		<-signCh
		log.Infof("Received signal, shutting down...")
		cancel()
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.Run(ctx); err != nil {
			log.Panic(err)
		}
	}()
	wg.Wait()
}
