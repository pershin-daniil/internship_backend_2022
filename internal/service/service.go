package service

import (
	"context"
	"fmt"

	"github.com/pershin-daniil/internship_backend_2022/internal/models"
	"github.com/sirupsen/logrus"
)

// Метод начисления средств на баланс. Принимает id пользователя и сколько средств зачислить.
// Метод резервирования средств с основного баланса на отдельном счете. Принимает id пользователя, ИД услуги, ИД заказа, стоимость.
// Метод признания выручки – списывает из резерва деньги, добавляет данные в отчет для бухгалтерии. Принимает id пользователя, ИД услуги, ИД заказа, сумму.
// Метод получения баланса пользователя. Принимает id пользователя.

type Store interface {
	AddFunds(ctx context.Context, data models.AddFundsRequest) (models.AddFundsRequest, error)
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

func (s *Service) AddFunds(ctx context.Context, data models.AddFundsRequest) (models.AddFundsRequest, error) {
	user, err := s.store.AddFunds(ctx, data)
	if err != nil {
		return models.AddFundsRequest{}, fmt.Errorf("service: %w", err)
	}
	return user, nil
}
