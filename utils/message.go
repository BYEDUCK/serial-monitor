package utils

import "time"

type Message struct {
	Timestamp time.Time
	Content   any
}

func NowMessage(msg any) *Message {
	return &Message{
		Timestamp: time.Now(),
		Content:   msg,
	}
}
