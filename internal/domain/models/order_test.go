package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrder_IsFinal(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{name: "order is processed", status: StatusProcessed, want: true},
		{name: "order is invalid", status: StatusInvalid, want: true},
		{name: "order status not in the list", status: "statusUnknown", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &Order{Status: tt.status}
			assert.Equal(t, tt.want, order.IsFinal())
		})
	}
}

func TestOrder_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		orderID string
		want    bool
	}{
		{name: "valid number", orderID: "6799837", want: true},
		{name: "invalid number", orderID: "679983", want: false},
		{name: "order id is not a number", orderID: "abc", want: false},
		{name: "empty order id", orderID: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &Order{OrderID: tt.orderID}
			assert.Equal(t, tt.want, order.IsValid())
		})
	}
}

func TestOrder_SetOrderID(t *testing.T) {
	type fields struct {
		OrderID        string
		AccrualOrderID string
	}
	type args struct {
		orderID string
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		wantOrderID string
	}{
		{name: "only order id is present", fields: fields{OrderID: "6799837"}, args: args{orderID: "6799837"}, wantOrderID: "6799837"},
		{name: "only accrual order id is present", fields: fields{AccrualOrderID: "6799837"}, args: args{orderID: "6799837"}, wantOrderID: "6799837"},
		{name: "both order id are present", fields: fields{AccrualOrderID: "123", OrderID: "6799837"}, wantOrderID: "123", args: args{orderID: "6799837"}},
		{name: "none are present", fields: fields{}, wantOrderID: "6799837", args: args{orderID: "6799837"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &Order{
				OrderID:        tt.fields.OrderID,
				AccrualOrderID: tt.fields.AccrualOrderID,
			}
			order.SetOrderID(tt.args.orderID)
			assert.Equal(t, tt.wantOrderID, order.OrderID)
		})
	}
}
