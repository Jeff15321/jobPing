package user

import "time"

type User struct {
	ID          int       `json:"id"`
	Email       string    `json:"email"`
	Preferences string    `json:"preferences"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
