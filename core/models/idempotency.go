package models

import "encoding/json"

type IdempotencyKey struct {
	UserID     string          `json:"user_id"`
	StatusCode int             `json:"status_code,omitempty"`
	Response   json.RawMessage `json:"response,omitempty"`
}
