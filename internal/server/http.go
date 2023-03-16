package server

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Server struct {
	log     *logrus.Entry
	address string
	version string
	server  *http.Server
}

func New(log *logrus.Logger, address string, version string) *Server {
	s := Server{
		log:     log.WithField("module", "server"),
		address: address,
		version: version,
	}
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/", s.rootHandler)
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

func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome"))
}
