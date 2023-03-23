package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pershin-daniil/internship_backend_2022/internal/logger"
	"github.com/pershin-daniil/internship_backend_2022/internal/models"
	"github.com/pershin-daniil/internship_backend_2022/internal/pgstore"
	"github.com/pershin-daniil/internship_backend_2022/internal/server"
	"github.com/pershin-daniil/internship_backend_2022/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
	"time"
)

var pgDSN = "postgres://postgres:secret@localhost:6432/internship?sslmode=disable"

const (
	port    = ":8080"
	version = "test"
	testURL = "http://localhost" + port
)

const (
	addFundsEndpoint = "/api/v1/addFunds"
)

type IntegrationTestSuite struct {
	suite.Suite
	log    *logrus.Logger
	store  *pgstore.Store
	app    *service.Service
	server *server.Server
	user   models.AddFundsRequest
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.log = logger.New()
	ctx := context.Background()
	var err error
	s.store, err = pgstore.New(ctx, s.log, pgDSN)
	s.Require().NoError(err)
	s.app = service.New(s.log, s.store)
	s.server = server.New(s.log, port, version, s.app)
	go func() {
		_ = s.server.Run(ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	err = s.store.ResetTables(ctx, []string{"history", "users"})
	s.Require().NoError(err)

}

func (s *IntegrationTestSuite) SetupTest() {
	s.user = s.createNewUserData()
}

func (s *IntegrationTestSuite) TestAddFunds() {
	s.T().Log(uuid.NewString())
	s.Run("addFunds for new user", func() {
		ctx := context.Background()
		var respUser models.AddFundsRequest
		resp := s.sendRequest(ctx, http.MethodPost, addFundsEndpoint, s.user, &respUser)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.user.UserID, respUser.UserID)
		s.Require().Equal(s.user.Balance, respUser.Balance)
	})

	s.Run("addFunds for old user", func() {
		s.user.TransactionID = uuid.NewString()
		ctx := context.Background()
		var respUser models.AddFundsRequest
		resp := s.sendRequest(ctx, http.MethodPost, addFundsEndpoint, s.user, &respUser)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.user.UserID, respUser.UserID)
		s.Require().Equal(s.user.Balance*2, respUser.Balance)
	})

	s.Run("addFunds for already added transaction", func() {
		ctx := context.Background()
		var respUser models.AddFundsRequest
		resp := s.sendRequest(ctx, http.MethodPost, addFundsEndpoint, s.user, &respUser)
		s.Require().Equal(http.StatusGone, resp.StatusCode)
		s.Require().Equal(s.user.Balance*2, 200)
	})
}

func (s *IntegrationTestSuite) sendRequest(ctx context.Context, method, endpoint string, body interface{}, dest interface{}) *http.Response {
	s.T().Helper()
	reqBody, err := json.Marshal(body)
	s.Require().NoError(err)
	req, err := http.NewRequestWithContext(ctx, method, testURL+endpoint, bytes.NewReader(reqBody))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer func() {
		err = resp.Body.Close()
		s.Require().NoError(err)
	}()
	if dest != nil {
		err = json.NewDecoder(resp.Body).Decode(&dest)
		s.Require().NoError(err)
	}
	return resp
}

func (s *IntegrationTestSuite) createNewUserData() models.AddFundsRequest {
	s.T().Helper()
	return models.AddFundsRequest{
		TransactionID: uuid.NewString(),
		UserID:        uuid.ClockSequence(),
		Balance:       100,
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
