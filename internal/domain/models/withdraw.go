package models

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type Withdraw struct {
	OrderID     OrderID   `json:"order" db:"order_id"`
	Sum         float64   `json:"sum" db:"sum"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
	Username    string    `json:"username" db:"username"`
}

func (w *Withdraw) ToNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{"order_id": w.OrderID, "username": w.Username, "processed_at": w.ProcessedAt, "sum": w.Sum}
}
