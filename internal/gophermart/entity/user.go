package entity

type Token string

type User struct {
	ID        int64
	Login     string
	Password  string
	Balance   float32
	Withdrawn float32
}
