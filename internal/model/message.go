package model

import "time"

type Message struct {
	Sender    string
	Body      string
	CreatedAt time.Time
}
