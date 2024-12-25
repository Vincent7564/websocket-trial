package models

type Chat struct {
	Content  string `json:"content" db:"content"`
	Username string `json:"username" db:"username"`
}
