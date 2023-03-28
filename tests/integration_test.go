package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/pershin-daniil/internship_backend_2022/pkg/models"
	"github.com/pershin-daniil/internship_backend_2022/pkg/pgstore"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pershin-daniil/internship_backend_2022/internal/logger"
	"github.com/pershin-daniil/internship_backend_2022/internal/server"
	"github.com/pershin-daniil/internship_backend_2022/pkg/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

var pgDSN = "postgres://postgres:secret@localhost:6432/internship?sslmode=disable"

const (
	port    = ":8080"
	version = "test"
	testURL = "http://localhost" + port
)

const (
	addFundsEndpoint         = "/api/v1/addFunds"
	reserveFundsEndpoint     = "/api/v1/reserveFunds"
	recognizeRevenueEndpoint = "/api/v1/recognizeRevenue"
	getWalletBalanceEndpoint = "/api/v1/getUserBalance"
)

type IntegrationTestSuite struct {
	suite.Suite
	log    *logrus.Logger
	store  *pgstore.Store
	app    *service.Service
	server *server.Server
	models.AddFundsRequest
	models.ReservedFundsRequest
	models.RecognizeRevenueRequest
	models.BalanceRequest
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
	err = s.store.ResetTables(ctx, []string{"events", "wallets"})
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) SetupTest() {
	s.AddFundsRequest = models.AddFundsRequest{
		TransactionID: uuid.NewString(),
		UserID:        1234,
		Balance:       100,
	}
	s.ReservedFundsRequest = models.ReservedFundsRequest{
		TransactionID: uuid.NewString(),
		WalletID:      1,
		ServiceID:     1,
		OrderID:       1111,
		Price:         10,
	}
	s.RecognizeRevenueRequest = models.RecognizeRevenueRequest{
		TransactionID: uuid.NewString(),
		WalletID:      1,
		ServiceID:     1,
		OrderID:       1111,
		Status:        "DONE",
	}
	s.BalanceRequest = models.BalanceRequest{UserID: 1234}
}

func (s *IntegrationTestSuite) TestMainWorkFlow() {
	s.Run("addFunds for new user", func() {
		ctx := context.Background()
		var respData models.WalletResponse
		resp := s.sendRequest(ctx, http.MethodPost, addFundsEndpoint, s.AddFundsRequest, &respData)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.AddFundsRequest.UserID, respData.UserID)
		s.Require().Equal(s.AddFundsRequest.Balance, respData.Balance)
	})

	s.Run("getBalance normal case", func() {
		ctx := context.Background()
		var respData models.WalletResponse
		resp := s.sendRequest(ctx, http.MethodGet, getWalletBalanceEndpoint, s.BalanceRequest, &respData)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.BalanceRequest.UserID, respData.UserID)
		s.Require().Equal(100, respData.Balance)
		s.Require().Equal(0, respData.Reserved)
	})

	s.Run("addFunds for old user", func() {
		s.AddFundsRequest.TransactionID = uuid.NewString()
		ctx := context.Background()
		var respData models.WalletResponse
		resp := s.sendRequest(ctx, http.MethodPost, addFundsEndpoint, s.AddFundsRequest, &respData)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.AddFundsRequest.UserID, respData.UserID)
		s.Require().Equal(s.AddFundsRequest.Balance*2, respData.Balance)
	})

	s.Run("getBalance normal case 2", func() {
		ctx := context.Background()
		var respData models.WalletResponse
		resp := s.sendRequest(ctx, http.MethodGet, getWalletBalanceEndpoint, s.BalanceRequest, &respData)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.BalanceRequest.UserID, respData.UserID)
		s.Require().Equal(200, respData.Balance)
		s.Require().Equal(0, respData.Reserved)
	})

	s.Run("addFunds for already added transaction", func() {
		ctx := context.Background()
		var respUser models.WalletResponse
		resp := s.sendRequest(ctx, http.MethodPost, addFundsEndpoint, s.AddFundsRequest, &respUser)
		s.Require().Equal(http.StatusGone, resp.StatusCode)
		s.Require().Equal(s.AddFundsRequest.Balance*2, 200)
	})

	s.Run("reserveFunds normal case", func() {
		ctx := context.Background()
		var respData models.EventsBodyResponse
		resp := s.sendRequest(ctx, http.MethodPost, reserveFundsEndpoint, s.ReservedFundsRequest, &respData)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.ReservedFundsRequest.WalletID, respData.WalletID)
		s.Require().Equal(s.ReservedFundsRequest.ServiceID, respData.ServiceID)
		s.Require().Equal(s.ReservedFundsRequest.OrderID, respData.OrderID)
		s.Require().Equal(s.ReservedFundsRequest.Price, respData.Price)
	})

	s.Run("getBalance with reserve", func() {
		ctx := context.Background()
		var respData models.WalletResponse
		resp := s.sendRequest(ctx, http.MethodGet, getWalletBalanceEndpoint, s.BalanceRequest, &respData)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.BalanceRequest.UserID, respData.UserID)
		s.Require().Equal(200, respData.Balance)
		s.Require().Equal(s.ReservedFundsRequest.Price, respData.Reserved)
	})

	s.Run("reserveFunds one more order", func() {
		ctx := context.Background()
		var respData models.EventsBodyResponse
		s.ReservedFundsRequest.TransactionID = uuid.NewString()
		s.ReservedFundsRequest.OrderID = 2222
		s.ReservedFundsRequest.Price = 50
		resp := s.sendRequest(ctx, http.MethodPost, reserveFundsEndpoint, s.ReservedFundsRequest, &respData)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.ReservedFundsRequest.WalletID, respData.WalletID)
		s.Require().Equal(s.ReservedFundsRequest.ServiceID, respData.ServiceID)
		s.Require().Equal(s.ReservedFundsRequest.OrderID, respData.OrderID)
		s.Require().Equal(s.ReservedFundsRequest.Price, respData.Price)
	})

	s.Run("getBalance with reserve 2", func() {
		ctx := context.Background()
		var respData models.WalletResponse
		resp := s.sendRequest(ctx, http.MethodGet, getWalletBalanceEndpoint, s.BalanceRequest, &respData)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.BalanceRequest.UserID, respData.UserID)
		s.Require().Equal(200, respData.Balance)
		s.Require().Equal(60, respData.Reserved)
	})

	s.Run("reserveFunds for already added transaction", func() {
		ctx := context.Background()
		var respData models.EventsBodyResponse
		resp := s.sendRequest(ctx, http.MethodPost, reserveFundsEndpoint, s.ReservedFundsRequest, &respData)
		s.Require().Equal(http.StatusGone, resp.StatusCode)
	})

	s.Run("reserveFunds same order", func() {
		s.ReservedFundsRequest.TransactionID = uuid.NewString()
		ctx := context.Background()
		var respData models.EventsBodyResponse
		resp := s.sendRequest(ctx, http.MethodPost, reserveFundsEndpoint, s.ReservedFundsRequest, &respData)
		s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("reserveFunds too high price", func() {
		s.T().Log(s.ReservedFundsRequest)
		s.ReservedFundsRequest.TransactionID = uuid.NewString()
		s.ReservedFundsRequest.OrderID = uuid.ClockSequence() + 10
		s.ReservedFundsRequest.Price = 100000
		ctx := context.Background()
		var respData models.EventsBodyResponse
		resp := s.sendRequest(ctx, http.MethodPost, reserveFundsEndpoint, s.ReservedFundsRequest, &respData)
		s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("recognizeRevenue normal case - status DONE", func() {
		ctx := context.Background()
		var respData models.EventsBodyResponse
		resp := s.sendRequest(ctx, http.MethodPost, recognizeRevenueEndpoint, s.RecognizeRevenueRequest, &respData)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.RecognizeRevenueRequest.Status, respData.Status)
		s.Require().Equal(s.RecognizeRevenueRequest.OrderID, respData.OrderID)
		s.Require().Equal(s.RecognizeRevenueRequest.ServiceID, respData.ServiceID)
		s.Require().Equal(s.RecognizeRevenueRequest.WalletID, respData.WalletID)
	})

	s.Run("recognizeRevenue same transaction", func() {
		ctx := context.Background()
		var respData models.EventsBodyResponse
		resp := s.sendRequest(ctx, http.MethodPost, recognizeRevenueEndpoint, s.RecognizeRevenueRequest, &respData)
		s.Require().Equal(http.StatusGone, resp.StatusCode)
	})

	s.Run("recognizeRevenue normal case - status CANCEL", func() {
		s.RecognizeRevenueRequest.TransactionID = uuid.NewString()
		s.RecognizeRevenueRequest.Status = "CANCELED"
		s.RecognizeRevenueRequest.OrderID = 2222
		ctx := context.Background()
		var respData models.EventsBodyResponse
		resp := s.sendRequest(ctx, http.MethodPost, recognizeRevenueEndpoint, s.RecognizeRevenueRequest, &respData)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().Equal(s.RecognizeRevenueRequest.Status, respData.Status)
		s.Require().Equal(s.RecognizeRevenueRequest.OrderID, respData.OrderID)
		s.Require().Equal(s.RecognizeRevenueRequest.ServiceID, respData.ServiceID)
		s.Require().Equal(s.RecognizeRevenueRequest.WalletID, respData.WalletID)
	})

	s.Run("recognizeRevenue weird order", func() {
		s.RecognizeRevenueRequest.TransactionID = uuid.NewString()
		s.RecognizeRevenueRequest.OrderID = 0
		ctx := context.Background()
		var respData models.EventsBodyResponse
		resp := s.sendRequest(ctx, http.MethodPost, recognizeRevenueEndpoint, s.RecognizeRevenueRequest, &respData)
		s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
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

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
