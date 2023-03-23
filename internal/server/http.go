package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

type Server struct {
	log     *logrus.Entry
	address string
	version string
	server  *http.Server
	app     App
}

func New(log *logrus.Logger, address string, version string, app App) *Server {
	s := Server{
		log:     log.WithField("module", "server"),
		address: address,
		version: version,
		app:     app,
	}
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/", s.rootHandler)
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Post("/addFunds", s.addFundsHandler)
			r.Post("/reserveFunds", s.reserveFundsHandler)
			r.Post("/recognizeRevenue", s.recognizeRevenueHandler)
			r.Get("/getUserBalance", s.getUserBalance)
		})
	})
	s.server = &http.Server{
		Addr:              s.address,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return &s
}

func (s *Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		gfCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		//nolint:contextcheck
		if err := s.server.Shutdown(gfCtx); err != nil {
			s.log.Warnf("err shutting down properly")
		}
	}()
	s.log.Infof("starting server on %s", s.address)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
