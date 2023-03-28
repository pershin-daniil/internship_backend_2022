package pgstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/pershin-daniil/internship_backend_2022/pkg/models"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

const retries = 3

var (
	ErrNotEnoughFunds    = fmt.Errorf("not enough funds")
	ErrUserNotExists     = fmt.Errorf("user doesn't exist")
	ErrOrderAlreadyAdded = fmt.Errorf("order has already added")
	ErrOrderNotExists    = fmt.Errorf("order doesn't exist")
)

type Store struct {
	log *logrus.Entry
	db  *sqlx.DB
}

func New(ctx context.Context, log *logrus.Logger, dsn string) (*Store, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("create new strore faild: %w", err)
	}
	return &Store{
		log: log.WithField("module", "pgstore"),
		db:  db,
	}, nil
}

func (s *Store) AddFunds(ctx context.Context, data models.AddFundsRequest) (models.AddFundsResponse, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return models.AddFundsResponse{}, fmt.Errorf("add funds faild: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.log.Warnf("add funds faild: %v", err)
		}
	}()

	var query strings.Builder

	query.WriteString(`INSERT INTO wallets (user_id, account_balance)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET
				account_balance = wallets.account_balance + $2,
				updated_at = NOW()
RETURNING id, user_id, account_balance, reserved, updated_at;`)

	var result models.AddFundsResponse

	for i := 0; i < retries; i++ {
		if err = tx.GetContext(ctx, &result, query.String(), data.UserID, data.Balance); err != nil {
			continue
		}
		if err = tx.Commit(); err != nil {
			return models.AddFundsResponse{}, err
		}
		return result, nil
	}

	return models.AddFundsResponse{}, fmt.Errorf("add funds faild: %w", err)
}

func (s *Store) ReserveFunds(ctx context.Context, data models.ReservedFundsRequest) (models.EventsBodyResponse, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return models.EventsBodyResponse{}, fmt.Errorf("reserve funds faild: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.log.Warnf("reserve funds faild: %v", err)
		}
	}()

	if ok, e := s.isEnoughFunds(ctx, tx, data.WalletID, data.Price); !ok {
		if e != nil {
			return models.EventsBodyResponse{}, fmt.Errorf("reserve funds faild: %w", e)
		}
		return models.EventsBodyResponse{}, ErrNotEnoughFunds
	}
	if ok, e := s.reserveFunds(ctx, tx, data.WalletID, data.Price); !ok {
		if e != nil {
			return models.EventsBodyResponse{}, fmt.Errorf("reserve funds faild: %w", e)
		}
		return models.EventsBodyResponse{}, ErrUserNotExists
	}

	query := `INSERT INTO events (wallet_id, service_id, order_id, price)
VALUES ($1, $2, $3, $4)
ON CONFLICT (order_id) DO NOTHING RETURNING id, wallet_id, service_id, order_id, price, datetime;`
	var result models.EventsBodyResponse

	for i := 0; i < retries; i++ {
		err = tx.GetContext(ctx, &result, query, data.WalletID, data.ServiceID, data.OrderID, data.Price)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.EventsBodyResponse{}, ErrOrderAlreadyAdded
		case err != nil:
			continue
		}
		if err = tx.Commit(); err != nil {
			return models.EventsBodyResponse{}, fmt.Errorf("reserved funds faild: %w", err)
		}
		return result, nil
	}

	return models.EventsBodyResponse{}, fmt.Errorf("reserved funds faild: %w", err)
}

func (s *Store) RecognizeRevenue(ctx context.Context, data models.RecognizeRevenueRequest) (models.EventsBodyResponse, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return models.EventsBodyResponse{}, fmt.Errorf("recognize revenue faild: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && errors.Is(err, sql.ErrTxDone) {
			s.log.Warnf("recognize revenue faild: %v", err)
		}
	}()
	price, err := s.checkPrice(ctx, tx, data.OrderID)
	if err != nil {
		return models.EventsBodyResponse{}, fmt.Errorf("recognize revenue faild: %w", err)
	}
	if err = s.changeBalance(ctx, tx, data.WalletID, price, data.Status); err != nil {
		return models.EventsBodyResponse{}, fmt.Errorf("recognize revenue faild: %w", err)
	}
	query := `
UPDATE events
SET status = $2
WHERE order_id = $1
RETURNING id, wallet_id, service_id, order_id, price, status, datetime`
	var result models.EventsBodyResponse
	for i := 0; i < retries; i++ {
		err = tx.GetContext(ctx, &result, query, data.OrderID, data.Status)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.EventsBodyResponse{}, fmt.Errorf("recognize revenue faild: %w", err)
		case err != nil:
			continue
		}
		if err = tx.Commit(); err != nil {
			return models.EventsBodyResponse{}, fmt.Errorf("recognize revenue faild: %w", err)
		}
		return result, nil
	}
	return models.EventsBodyResponse{}, fmt.Errorf("recognize revenue faild: %w", err)
}

type q interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func (s *Store) isEnoughFunds(ctx context.Context, q q, id int, price int) (bool, error) {
	query := `
SELECT account_balance - wallets.reserved AS balance
FROM wallets
WHERE id = $1;`
	var balance int
	var err error
	for i := 0; i < retries; i++ {
		err = q.GetContext(ctx, &balance, query, id)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, nil
		case err != nil:
			continue
		}
		return balance >= price, nil
	}
	return false, fmt.Errorf("check if enough funds faild: %w", err)
}

func (s *Store) reserveFunds(ctx context.Context, q q, id int, price int) (bool, error) {
	query := `
UPDATE wallets
SET reserved = reserved + $2
WHERE id = $1
RETURNING TRUE;`
	var ok bool
	var err error
	for i := 0; i < retries; i++ {
		err = q.GetContext(ctx, &ok, query, id, price)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, nil
		case err != nil:
			continue
		}
		return ok, nil
	}
	return false, fmt.Errorf("reserve faild: %w", err)
}

func (s *Store) changeBalance(ctx context.Context, q q, id int, price int, status string) error {
	var query string
	switch status {
	case "DONE":
		query = `UPDATE wallets
		SET account_balance = account_balance - $2,
			reserved = reserved - $2
		WHERE id = $1
		RETURNING TRUE;`
	case "CANCELED":
		query = `UPDATE wallets
		SET reserved = reserved - $2
		WHERE id = $1
		RETURNING TRUE;`
	}
	var ok bool
	var err error
	for i := 0; i < retries; i++ {
		if err = q.GetContext(ctx, &ok, query, id, price); err != nil {
			continue
		}
		if ok {
			return nil
		}
	}
	return fmt.Errorf("change balance faild: %v", err)
}

func (s *Store) checkPrice(ctx context.Context, q q, orderID int) (int, error) {
	query := `
SELECT price FROM events
WHERE order_id = $1;`
	var price int
	var err error
	for i := 0; i < retries; i++ {
		err = q.GetContext(ctx, &price, query, orderID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrOrderNotExists
		case err != nil:
			continue
		}
		return price, nil
	}
	return 0, fmt.Errorf("check price faild: %w", err)
}

func (s *Store) ResetTables(ctx context.Context, tables []string) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE TABLE`+` `+strings.Join(tables, `, `))
	for _, table := range tables {
		_, err = s.db.ExecContext(ctx, fmt.Sprintf(`ALTER SEQUENCE %s_id_seq RESTART`, table))
		if err != nil {
			return err
		}
	}
	return err
}
