package models

import "time"

type Team struct {
	ID        string    `json:"id" gorm:"type:uuid;default:gen_random_uuid()"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
