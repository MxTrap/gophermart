package entity

type Token string

type User struct {
	ID       int64
	Login    string
	Password string
}
