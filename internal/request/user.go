package request

import (
	"errors"
)

var (
	FullNameIsEmpty = errors.New("FullName can't be empty")
)

type User struct {
	FullName string `json:"fullName"`
}

func (u *User) Valid() error {
	if len(u.FullName) == 0 {
		return FullNameIsEmpty
	}
	return nil
}
