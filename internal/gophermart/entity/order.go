package entity

type Order struct {
	userID     int64
	Number     string
	Status     string
	Accrual    int
	UploadedAt int
}
