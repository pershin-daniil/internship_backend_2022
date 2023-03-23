package models

type AddFundsRequest struct {
	TransactionID string `json:"transactionID" db:"_"`
	UserID        int    `json:"userID" db:"user_id"`
	Balance       int    `json:"account_balance" db:"account_balance"`
}
