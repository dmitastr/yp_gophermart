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
	OrderId        string    `json:"number" db:"order_id"`
	AccrualOrderId string    `json:"order,omitempty" db:"-"`
	Status         string    `json:"status" db:"status"`
	Accrual        float64   `json:"accrual" db:"accrual"`
	UploadedAt     time.Time `json:"uploaded_at" db:"uploaded_at"`
	Username       string    `json:"username" db:"username"`
}

func NewOrder(orderId string, username string) *Order {
	return &Order{OrderId: orderId, UploadedAt: time.Now(), Username: username}
}

func (order *Order) ToNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{"order_id": order.OrderId, "username": order.Username, "uploaded_at": order.UploadedAt,
		"accrual": order.Accrual, "status": order.Status}

}

func (order *Order) IsFinal() bool {
	return order.Status == StatusProcessed || order.Status == StatusInvalid
}

func (order *Order) SetOrderId(orderId string) {
	accrualNumber := order.AccrualOrderId
	order.AccrualOrderId = ""
	if accrualNumber != "" {
		order.OrderId = accrualNumber
		return
	}
	order.OrderId = orderId
}
