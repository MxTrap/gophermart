package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtService struct {
	secretKey string
}

func NewJWTService(secretKey string) *JwtService {
	return &JwtService{
		secretKey: secretKey,
	}
}

func (s JwtService) GenerateAccessToken(user entity.User, ttl time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.Id
	claims["login"] = user.Login
	claims["exp"] = time.Now().Add(ttl).Unix()

	signedString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return signedString, nil
}

func (s JwtService) GenerateRefreshToken(userId int64) string {
	return fmt.Sprintf("%s-%d", uuid.New().String(), userId)
}

func (s JwtService) GetUserIdFromRefreshToken(token entity.Token) (int64, error) {
	index := strings.LastIndex(string(token), "_")
	userId, err := strconv.ParseInt(string(token)[index:], 10, 64)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func (s JwtService) Parse(token entity.Token) (int64, error) {
	parsedToken, err := jwt.Parse(string(token), func(token *jwt.Token) (any, error) {
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return 0, err
	}

	if !parsedToken.Valid {
		return 0, common.ErrInvalidToken
	}

	expTime, err := parsedToken.Claims.GetExpirationTime()

	if err != nil {
		return 0, err
	}

	if expTime.Unix() < time.Now().Unix() {
		return 0, common.ErrInvalidToken
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, common.ErrInvalidToken
	}
	uid, ok := claims["uid"]
	if !ok {
		return 0, common.ErrInvalidToken
	}

	fuid, ok := uid.(float64)

	if !ok {
		return 0, common.ErrInvalidToken
	}

	return int64(fuid), nil
}
