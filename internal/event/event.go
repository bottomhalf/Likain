package event

import "encoding/json"

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Handler func(Event, ClientIface) error

type ClientIface interface {
	Send(Event)
}
