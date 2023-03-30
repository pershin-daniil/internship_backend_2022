package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pershin-daniil/internship_backend_2022/pkg/models"
	"github.com/pershin-daniil/internship_backend_2022/pkg/pgstore"
)

var txs = make(map[string]struct{})

type App interface {
	AddFunds(ctx context.Context, data models.AddFundsRequest) (models.WalletResponse, error)
	WalletBalance(ctx context.Context, data models.BalanceRequest) (models.WalletResponse, error)
	ReserveFunds(ctx context.Context, data models.ReservedFundsRequest) (models.EventsBodyResponse, error)
	RecognizeRevenue(ctx context.Context, data models.RecognizeRevenueRequest) (models.EventsBodyResponse, error)
}

func (s *Server) addFundsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var data models.AddFundsRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}
	if _, ok := txs[data.TransactionID]; ok {
		s.writeResponse(w, http.StatusConflict, nil)
		return
	}
	resp, err := s.app.AddFunds(ctx, data)
	if err != nil {
		s.log.Warnf("err during add funds: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	txs[data.TransactionID] = struct{}{}
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
		s.writeResponse(w, http.StatusConflict, nil)
		return
	}
	resp, err := s.app.ReserveFunds(ctx, data)
	switch {
	case errors.Is(err, pgstore.ErrOrderAlreadyAdded):
		s.log.Warnf("err during reserve funds: %v", err)
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	case errors.Is(err, pgstore.ErrNotEnoughFunds):
		s.log.Warnf("err during reserve funds: %v", err)
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	case err != nil:
		s.log.Warnf("err during reserve funds: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}
	txs[data.TransactionID] = struct{}{}
	s.writeResponse(w, http.StatusOK, resp)
}

func (s *Server) recognizeRevenueHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var data models.RecognizeRevenueRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil)
		return
	}
	if _, ok := txs[data.TransactionID]; ok {
		s.writeResponse(w, http.StatusConflict, nil)
		return
	}
	resp, err := s.app.RecognizeRevenue(ctx, data)
	switch {
	case errors.Is(err, pgstore.ErrOrderNotExists):
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	case err != nil:
		s.log.Warnf("err during recognize revenue: %v", err)
		s.writeResponse(w, http.StatusInternalServerError, err)
	}
	txs[data.TransactionID] = struct{}{}
	s.writeResponse(w, http.StatusOK, resp)
}

func (s *Server) getUserBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var data models.BalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil)
		return
	}
	resp, err := s.app.WalletBalance(ctx, data)
	switch {
	case errors.Is(err, pgstore.ErrUserNotExists):
		s.writeResponse(w, http.StatusBadRequest, err)
	case err != nil:
		s.log.Warnf("err during getting balance (id %d): %v", data.UserID, err)
		return
	}
	s.writeResponse(w, http.StatusOK, resp)
}

func (s *Server) writeResponse(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if x, ok := data.(error); ok {
		if err := json.NewEncoder(w).Encode(ErrorResponse{Error: x.Error()}); err != nil {
			s.log.Warnf("write response failed: %v", err)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Warnf("write response failed: %v", err)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}
