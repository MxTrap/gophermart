package entity

type Order struct {
	UserID     int64
	Number     string
	Status     string
	Accrual    int
	UploadedAt int
}
