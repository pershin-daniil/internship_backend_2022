package models

type AddFundsRequest struct {
	TransactionID string `json:"transactionID" db:"_"`
	UserID        int    `json:"userID" db:"user_id"`
	Balance       int    `json:"account_balance" db:"account_balance"`
}

type ReservedFundsRequest struct {
	TransactionID string `json:"transactionID" db:"_"`
	UserID        int    `json:"userID" db:"user_id"`
	ServiceID     int    `json:"serviceID" db:"service_id"`
	OrderID       int    `json:"orderID" db:"order_id"`
	Price         int    `json:"price" db:"price"`
}
