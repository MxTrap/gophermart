package entity

import "time"

const (
	OrderNew        = "NEW"
	OrderProcessing = "PROCESSING"
	OrderInvalid    = "INVALID"
	OrderProcessed  = "PROCESSED"
)

type Order struct {
	UserID     int64
	Number     string
	Status     string
	Accrual    *float32
	UploadedAt time.Time
}
