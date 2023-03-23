package pgstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pershin-daniil/internship_backend_2022/internal/models"
	"github.com/sirupsen/logrus"
)

const retries = 3

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

func (s *Store) AddFunds(ctx context.Context, data models.AddFundsRequest) (models.AddFundsRequest, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return models.AddFundsRequest{}, fmt.Errorf("add funds faild: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.log.Warnf("add funds faild: %v", err)
		}
	}()

	var query strings.Builder
	exist, err := s.isUserExist(ctx, tx, data.UserID)
	if err != nil {
		return models.AddFundsRequest{}, fmt.Errorf("add funds faild: %w", err)
	}
	if !exist {
		query.WriteString(`
INSERT INTO users (user_id, account_balance)
VALUES ($1, $2)
RETURNING user_id, account_balance;`)
	} else {
		query.WriteString(`
UPDATE users
SET account_balance = account_balance + $2
WHERE user_id = $1
RETURNING user_id, account_balance;`)
	}
	var result models.AddFundsRequest

	for i := 0; i < retries; i++ {
		if err = s.db.GetContext(ctx, &result, query.String(), data.UserID, data.Balance); err != nil {
			continue
		}
		return result, nil
	}
	return models.AddFundsRequest{}, fmt.Errorf("add funds faild: %w", err)
}

type q interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func (s *Store) isUserExist(ctx context.Context, q q, id int) (bool, error) {
	query := `
SELECT TRUE
FROM users
WHERE user_id = $1;`
	var exist bool
	var err error
	for i := 0; i < retries; i++ {
		err = q.GetContext(ctx, &exist, query, id)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, nil
		case err != nil:
			continue
		}
		return exist, nil
	}
	return false, fmt.Errorf("check if user exists faild: %w", err)
}

func (s *Store) ResetTables(ctx context.Context, tables []string) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE TABLE`+` `+strings.Join(tables, `, `))
	return err
}
