package entity

type Token string

type User struct {
	Id       int64
	Login    string
	Password string
}
