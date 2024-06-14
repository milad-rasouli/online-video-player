package request

import (
	"errors"
)

type Message struct {
	Sender string `json:"sender"`
	Body   string `json:"body"`
}

var (
	InvalidSenderSize = errors.New("Sender must be between 1 and 100")
	InvalidBodySize   = errors.New("Body must be between 1 and 500")
)

func (m Message) Validate() error {
	if len(m.Sender) == 0 || len(m.Sender) > 100 {
		return InvalidSenderSize
	}
	if len(m.Body) == 0 || len(m.Body) > 500 {
		return InvalidBodySize
	}
	return nil
}
