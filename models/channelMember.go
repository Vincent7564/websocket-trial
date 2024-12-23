package models

import "time"

type ChannelMember struct {
	UserID    int       `json:"user_id" db:"user_id"`
	ChannelID int       `json:"channel_id" db:"channel_id"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
}
