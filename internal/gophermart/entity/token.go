package entity

type Token string

type RefreshToken struct {
	Id             int64 `db:"id"`
	RefreshToken   Token `db:"refresh_token"`
	ExpirationTime int64 `db:"expiration_time"`
	UserId         int64 `db:"user_id"`
}

type Tokens struct {
	AccessToken  Token
	RefreshToken Token
}
