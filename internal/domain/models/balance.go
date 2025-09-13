package models

type Balance struct {
	Username  string  `json:"-"`
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (b *Balance) CanWithdraw(sum float64) bool {
	return b.Current >= sum
}
