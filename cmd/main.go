package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pershin-daniil/internship_backend_2022/internal/logger"
	"github.com/pershin-daniil/internship_backend_2022/internal/pgstore"
	"github.com/pershin-daniil/internship_backend_2022/internal/server"
	"github.com/pershin-daniil/internship_backend_2022/internal/service"
)

const (
	address = ":8080"
	version = "0.0.1"
)

var pgDSN = os.Getenv("PG_DSN")

func main() {
	log := logger.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := pgstore.New(ctx, log, pgDSN)
	if err != nil {
		log.Panic(err)
	}

	app := service.New(log, store)

	s := server.New(log, address, version, app)

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
