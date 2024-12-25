package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"
	"websocket-trial/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

type PaginateData struct {
	Limit int32
	Page  int32
	Total int32
}

type Response struct {
	Message  string
	Data     interface{}   `json:"data"`
	Paginate *PaginateData `json:"paginate,omitempty"`
}

var validate *validator.Validate

type ErrorResponse struct {
	Error       bool
	FailedField string
	Tag         string
	Value       interface{}
}

func GenerateResponse(ctx *fiber.Ctx, statusCode int, respmsg string, result interface{}) error {
	resp := Response{
		Message:  respmsg,
		Data:     result,
		Paginate: nil,
	}
	return ctx.Status(statusCode).JSON(resp)

}

func GenerateResponsePaginate(ctx *fiber.Ctx, statusCode int, respmsg string, result interface{}, paginate PaginateData) error {
	resp := Response{
		Message:  respmsg,
		Data:     result,
		Paginate: &paginate,
	}
	return ctx.Status(statusCode).JSON(resp)

}

func Validate(data interface{}) []ErrorResponse {
	validationErrors := []ErrorResponse{}
	validate = validator.New()
	errs := validate.Struct(data)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			var elem ErrorResponse

			elem.FailedField = err.Field()
			elem.Tag = err.Tag()
			elem.Value = err.Value()
			elem.Error = true

			validationErrors = append(validationErrors, elem)
		}
	}

	return validationErrors
}

func ValidateData(data interface{}) []string {
	var errorMessage []string
	listError := Validate(data)
	if len(listError) > 0 && listError[0].Error {
		for _, err := range listError {
			errorMessage = append(errorMessage, fmt.Sprintf(
				"%s: '%v' | Needs to implement '%s'",
				err.FailedField,
				err.Value,
				err.Tag,
			))
		}
	}
	return errorMessage
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CheckToken(db *gorm.DB, token string) bool {
	var entities models.UserAccessToken

	claims, err := DecodeToken(token)

	if err != nil {
		return false
	}
	fmt.Printf("Test 1")
	exp, ok := (*claims)["exp"].(float64)

	if !ok {
		return false
	}

	expirationTime := time.Unix(int64(exp), 0)

	if time.Now().After(expirationTime) {
		return false
	}

	if err := db.Where("token = ?", token).First(&entities).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
		return false
	}

	if time.Now().After(entities.ExpiredAt) {
		return false
	}

	return true
}

func DecodeToken(tokenString string) (*jwt.MapClaims, error) {
	var secret_token = os.Getenv("JWT_SECRET_KEY")
	print(tokenString)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret_token), nil
	})

	if err != nil {
		fmt.Printf("error Parsing Token " + err.Error())
		return nil, fmt.Errorf("error Parsing Token :%w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, fmt.Errorf("invalid Token")
}

func SendEmail(to string, subject string, content string) error {
	m := gomail.NewMessage()
	email := os.Getenv("USER_EMAIL")
	port := 587
	password := os.Getenv("USER_PASSWORD")
	if email == "" || password == "" {
		return fmt.Errorf("email credentials not set in environment variables")
	}
	m.SetHeader("From", email)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	d := gomail.NewDialer("smtp.gmail.com", port, email, password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func GenerateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
