package models

type User struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	// PasswordHash      string    `json:"-" db:"password_hash"`
	// CreatedAt         time.Time `json:"created_at" db:"created_at"`
	// Email             string    `json:"email,omitempty" db:"email"`
	// ProfilePictureURL string    `json:"profile_picture_url,omitempty" db:"profile_picture_url"`
}
