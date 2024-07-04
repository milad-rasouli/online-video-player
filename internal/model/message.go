package model

import "time"

type Message struct {
	Sender    string    `json:"sender"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}
