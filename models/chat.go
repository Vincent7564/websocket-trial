package models

type Chat struct {
	Content string `json:"content" db:"content"`
}
