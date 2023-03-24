package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/pershin-daniil/internship_backend_2022/internal/pgstore"
	"net/http"

	"github.com/pershin-daniil/internship_backend_2022/internal/models"
)

var txs = make(map[string]struct{})

type App interface {
	AddFunds(ctx context.Context, data models.AddFundsRequest) (models.AddFundsRequest, error)
	ReserveFunds(ctx context.Context, data models.ReservedFundsRequest) (models.ReservedFundsRequest, error)
}

func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("welcome"))
	if err != nil {
		s.writeResponse(w, http.StatusInternalServerError, err)
	}
}

func (s *Server) addFundsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var data models.AddFundsRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	if _, ok := txs[data.TransactionID]; ok {
		s.writeResponse(w, http.StatusGone, nil)
		return
	}
	txs[data.TransactionID] = struct{}{}
	resp, err := s.app.AddFunds(ctx, data)
	if err != nil {
		s.log.Warnf("err during add funds: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	resp.TransactionID = data.TransactionID
	s.writeResponse(w, http.StatusOK, resp)
}

func (s *Server) reserveFundsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var data models.ReservedFundsRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	if _, ok := txs[data.TransactionID]; ok {
		s.writeResponse(w, http.StatusGone, nil)
		return
	}
	txs[data.TransactionID] = struct{}{}
	s.log.Infof("%s", data.TransactionID)
	resp, err := s.app.ReserveFunds(ctx, data)
	switch {
	case errors.Is(err, pgstore.ErrOrderAlreadyAdded):
		s.log.Warnf("err during reserve funds: %v", err)
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	case err != nil:
		s.log.Warnf("err during reserve funds: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	resp.TransactionID = data.TransactionID
	s.writeResponse(w, http.StatusOK, resp)
}

func (s *Server) recognizeRevenueHandler(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) getUserBalance(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) writeResponse(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if x, ok := data.(error); ok {
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: x.Error()}); err != nil {
			s.log.Warnf("write response faild: %v", err)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Warnf("write response faild: %v", err)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}
