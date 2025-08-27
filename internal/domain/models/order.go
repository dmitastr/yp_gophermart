package models

import "time"

type Order struct {
	OrderId    string    `json:"number" db:"order_id"`
	Status     string    `json:"status" db:"status"`
	Accrual    float64   `json:"accrual" db:"accrual"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}
