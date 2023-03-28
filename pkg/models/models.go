package models

import "time"

type AddFundsRequest struct {
	TransactionID string `json:"transactionID"`
	UserID        int    `json:"userID" db:"user_id"`
	Balance       int    `json:"balance" db:"account_balance"`
}

type AddFundsResponse struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"userID" db:"user_id"`
	Balance   int       `json:"balance" db:"account_balance"`
	Reserved  int       `json:"reserved" db:"reserved"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type ReservedFundsRequest struct {
	TransactionID string `json:"transactionID"`
	WalletID      int    `json:"walletID" db:"wallet_id"`
	ServiceID     int    `json:"serviceID" db:"service_id"`
	OrderID       int    `json:"orderID" db:"order_id"`
	Price         int    `json:"price" db:"price"`
}

type RecognizeRevenueRequest struct {
	TransactionID string `json:"transactionID"`
	WalletID      int    `json:"walletID" db:"wallet_id"`
	ServiceID     int    `json:"serviceID" db:"service_id"`
	OrderID       int    `json:"orderID" db:"order_id"`
	Status        string `json:"status" db:"status"`
}

type EventsBodyResponse struct {
	ID        int       `json:"id" db:"id"`
	WalletID  int       `json:"walletID" db:"wallet_id"`
	ServiceID int       `json:"serviceID" db:"service_id"`
	OrderID   int       `json:"orderID" db:"order_id"`
	Price     int       `json:"price" db:"price"`
	Status    string    `json:"status" db:"status"`
	DateTime  time.Time `json:"dateTime" db:"datetime"`
}
