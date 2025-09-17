package models

import (
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	StatusProcessed  = "PROCESSED"
	StatusNew        = "NEW"
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
)

type Order struct {
	OrderID        OrderID   `json:"number" db:"order_id"`
	AccrualOrderID string    `json:"order,omitempty" db:"-"`
	Status         string    `json:"status" db:"status"`
	Accrual        float64   `json:"accrual" db:"accrual"`
	UploadedAt     time.Time `json:"uploaded_at" db:"uploaded_at"`
	Username       string    `json:"username" db:"username"`
}

func NewOrder(orderID string, username string) *Order {
	return &Order{OrderID: OrderID(orderID), UploadedAt: time.Now(), Username: username, Status: StatusNew}
}

func (order *Order) ToNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{"order_id": order.OrderID, "username": order.Username, "uploaded_at": order.UploadedAt,
		"accrual": order.Accrual, "status": order.Status}

}

func (order *Order) IsFinal() bool {
	return order.Status == StatusProcessed || order.Status == StatusInvalid
}

func (order *Order) IsValid() bool {
	return order.OrderID.IsValid()
}

func (order *Order) SetOrderID(orderID string) {
	accrualNumber := order.AccrualOrderID
	order.AccrualOrderID = ""
	if accrualNumber != "" {
		order.OrderID = OrderID(accrualNumber)
		return
	}
	order.OrderID = OrderID(orderID)
}
