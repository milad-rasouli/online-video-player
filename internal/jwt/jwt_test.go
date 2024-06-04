package jwt

import (
	"fmt"
	"testing"

	"github.com/Milad75Rasouli/online-video-player/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestAccessTokenStuff(t *testing.T) {
	var (
		token  string
		err    error
		secret = "verySecret"
	)
	user := model.User{
		FullName: "foo bar",
	}
	jwt := NewAccessJWT(secret)

	{
		token, err = jwt.Create(user)
		assert.NoError(t, err)
		fmt.Printf("created token: %s\n", token)
	}

	{
		jwtUser, err := jwt.VerifyParse(token)
		if err != nil {
			assert.NoError(t, err)
		}
		fmt.Printf("parsed token: %+v\n", jwtUser)

	}
	{
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImJhckBiYXouY29tIiwiZXhwIjoxNzE0MDc4MTYxLCJmdWxsX25hbWUiOiJmb28gYmFyIiwicm9sZSI6ImFkbWluIn0.pLMe3zPYggvKpA8SDL2mbV6kfhISW1IH-xgBClS40TI"
		jwtUser, err := jwt.VerifyParse(expiredToken)
		assert.Error(t, err)
		fmt.Printf("parsed token: %+v expiration error: %+v\n", jwtUser, err)
	}
}
