package jwt

import (
	"errors"
	"time"

	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenExpireAfter = 8 // Hours
)

var (
	InvalidJWTMethod   = errors.New("Invalid JWT Method")
	FailedToParseToken = errors.New("Failed to parse the token")
	InvalidToken       = errors.New("Invalid token")
	InvalidTokenClaims = errors.New("Invalid token claims")
	FailedToReadClaims = errors.New("Failed to read the claims")
)

type AccessJWT struct {
	SecretKey string
}

func NewAccessJWT(secret string) AccessJWT {
	return AccessJWT{SecretKey: secret}
}

func (j AccessJWT) Create(jwtUser model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"full_name": jwtUser.FullName,
		"exp":       time.Now().Add(time.Hour * AccessTokenExpireAfter).Unix(),
	})

	tokenString, err := token.SignedString([]byte(j.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Caution: the jwtUser "InitiateTime" won't effect the final result. it's always time.Now()
func (j AccessJWT) VerifyParse(tokenString string) (model.User, error) {
	var jwtUser model.User

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, InvalidJWTMethod
		}

		return []byte(j.SecretKey), nil
	})

	if err != nil {
		return jwtUser, errors.Join(FailedToParseToken, err)
	}
	if token.Valid == false {
		return jwtUser, InvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return jwtUser, InvalidTokenClaims
	}

	var hasItem bool
	jwtUser.FullName, hasItem = claims["full_name"].(string)
	if hasItem == false {
		return jwtUser, errors.Join(FailedToReadClaims, errors.New("full name claims error"))
	}
	return jwtUser, nil
}
