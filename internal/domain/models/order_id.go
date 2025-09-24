package models

import (
	"strconv"

	"github.com/theplant/luhn"
)

type OrderID string

func (orderID OrderID) IsValid() bool {
	orderIDInt, err := strconv.Atoi(string(orderID))
	if err != nil {
		return false
	}
	return luhn.Valid(orderIDInt)

}
