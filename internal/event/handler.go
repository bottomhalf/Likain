package event

import (
	"Confeet/internal/model"
	"encoding/json"
	"log"
)

const (
	EventSendMessage = "send_message"
	EventNewMessage  = "new_message"
)

type SendMessagePayload struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

func HandleSendMessage(ev Event, c ClientIface) error {
	resp := model.Reponse{
		Message: "Message recieved",
		From:    "server",
	}

	payloadBytes, _ := json.Marshal(resp)
	broadcastEvent := Event{
		Type:    EventNewMessage,
		Payload: json.RawMessage(payloadBytes),
	}

	log.Printf("New message from %s: %s\n", ev.Type, ev.Payload)

	c.Send(broadcastEvent)
	return nil
}
