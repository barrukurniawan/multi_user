package models

import "time"

type User struct {
	ID        string    `json:"id" gorm:"type:uuid;default:gen_random_uuid()"`
	TeamID    string    `json:"team_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
