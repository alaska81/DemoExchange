package apiservice

import (
	"encoding/json"
)

type Method string

const (
	GET  Method = "GET"
	POST Method = "POST"
)

func (m Method) String() string {
	return string(m)
}

type Response struct {
	Success bool            `json:"success"`
	Return  json.RawMessage `json:"return"`
	Error   string          `json:"error"`
}
