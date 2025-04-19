package entity

import "time"

type Balance struct {
	Current   float32
	Withdrawn float32
}

type Withdrawal struct {
	Order       string
	Sum         float32
	ProcessedAt time.Time
}
