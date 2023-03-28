package service

import (
	"context"
	"fmt"

	"github.com/pershin-daniil/internship_backend_2022/pkg/models"

	"github.com/sirupsen/logrus"
)

type Store interface {
	AddFunds(ctx context.Context, data models.AddFundsRequest) (models.WalletResponse, error)
	ReserveFunds(ctx context.Context, data models.ReservedFundsRequest) (models.EventsBodyResponse, error)
	RecognizeRevenue(ctx context.Context, data models.RecognizeRevenueRequest) (models.EventsBodyResponse, error)
	WalletBalance(ctx context.Context, data models.BalanceRequest) (models.WalletResponse, error)
}

type Service struct {
	log   *logrus.Entry
	store Store
}

func New(log *logrus.Logger, store Store) *Service {
	return &Service{
		log:   log.WithField("module", "service"),
		store: store,
	}
}

func (s *Service) AddFunds(ctx context.Context, data models.AddFundsRequest) (models.WalletResponse, error) {
	wallet, err := s.store.AddFunds(ctx, data)
	if err != nil {
		return models.WalletResponse{}, fmt.Errorf("service: %w", err)
	}
	return wallet, nil
}

func (s *Service) ReserveFunds(ctx context.Context, data models.ReservedFundsRequest) (models.EventsBodyResponse, error) {
	reserved, err := s.store.ReserveFunds(ctx, data)
	if err != nil {
		return models.EventsBodyResponse{}, fmt.Errorf("service: %w", err)
	}
	return reserved, nil
}

func (s *Service) RecognizeRevenue(ctx context.Context, data models.RecognizeRevenueRequest) (models.EventsBodyResponse, error) {
	recognized, err := s.store.RecognizeRevenue(ctx, data)
	if err != nil {
		return models.EventsBodyResponse{}, fmt.Errorf("service: %w", err)
	}
	return recognized, nil
}

func (s *Service) WalletBalance(ctx context.Context, data models.BalanceRequest) (models.WalletResponse, error) {
	balance, err := s.store.WalletBalance(ctx, data)
	if err != nil {
		return models.WalletResponse{}, fmt.Errorf("service: %w", err)
	}
	return balance, nil
}
