package entity

type Order struct {
	UserId     int64
	Number     string
	Status     string
	Accrual    int
	UploadedAt int
}
