package models

import "encoding/json"

type IdempotencyKey struct {
	RequestHash string          `json:"request_hash"`
	StatusCode  int             `json:"status_code,omitempty"`
	Response    json.RawMessage `json:"response,omitempty"`
}
