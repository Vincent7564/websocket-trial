package dto

type RegisterRequest struct {
	Username string `json:"username" gorm:"type:text;unique" validate:"omitempty"`
	Password string `json:"password" validate:"gte=8,lte=30"`
	Email    string `json:"email" gorm:"type:text;unique" validate:"email"`
}
