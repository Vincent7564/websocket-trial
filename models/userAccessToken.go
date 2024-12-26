package models

import "time"

type UserAccessToken struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	UserID    int    `json:"user_id" gorm:"type:int"`
	Token     string `json:"token" gorm:"type:text"`
	CreatedAt time.Time
	ExpiredAt time.Time
}
