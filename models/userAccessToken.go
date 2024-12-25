package models

import "time"

type UserAccessToken struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	UserID    uint   `json:"user_id" gorm:"type:uint"`
	Token     string `json:"token" gorm:"type:text"`
	CreatedAt time.Time
	ExpiredAt time.Time
}
