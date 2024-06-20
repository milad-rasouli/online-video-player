package jwt

import (
	"errors"
	"log"
	"time"

	"github.com/Milad75Rasouli/online-video-player/internal/config"
	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

var (
	InvalidJWTMethod   = errors.New("Invalid JWT Method")
	FailedToParseToken = errors.New("Failed to parse the token")
	InvalidToken       = errors.New("Invalid token")
	InvalidTokenClaims = errors.New("Invalid token claims")
	FailedToReadClaims = errors.New("Failed to read the claims")
)

type AccessJWT struct {
	cfg config.Config
}

func NewAccessJWT(cfg config.Config) AccessJWT {
	return AccessJWT{cfg: cfg}
}

func (j AccessJWT) Create(jwtUser model.User) (string, error) {
	t := time.Now().Add(time.Minute * time.Duration(j.cfg.JWtExpireTime)).Unix()
	log.Printf("expration time: %s\n", time.Unix(t, 0).String()) //TODO:delete this
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"full_name": jwtUser.FullName,
		"exp":       t,
	})

	tokenString, err := token.SignedString([]byte(j.cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j AccessJWT) VerifyParse(tokenString string) (model.User, error) {
	var jwtUser model.User

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, InvalidJWTMethod
		}

		return []byte(j.cfg.JWTSecret), nil
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
