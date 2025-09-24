package client

import (
	"Likain/internal/event"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	conn    *websocket.Conn
	manager ManagerIface
	egress  chan event.Event
}

type ManagerIface interface {
	RouteEvent(event.Event, *Client) error
	RemoveClient(*Client)
}

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

func New(conn *websocket.Conn, m ManagerIface) *Client {
	return &Client{
		conn:    conn,
		manager: m,
		egress:  make(chan event.Event),
	}
}

func (c *Client) ReadMessages() {
	defer c.manager.RemoveClient(c)

	c.conn.SetReadLimit(512)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}

	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}

		var ev event.Event
		if err := json.Unmarshal(msg, &ev); err != nil {
			log.Println("unmarshal error: ", err)
			break
		}

		if err := c.manager.RouteEvent(ev, c); err != nil {
			log.Printf("error handling message: %v", err)
		}
	}
}

func (c *Client) WriteMessage() {
	defer func() {
		c.manager.RemoveClient(c)
	}()

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case ev, ok := <-c.egress:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Printf("connection closed: %v", err)
				}
				return
			}

			data, err := json.Marshal(ev)
			if err != nil {
				log.Println("marshal error: ", err)
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println("failed to send messgage erro: ", err)
				return
			}

			log.Println("message send")

		case <-ticker.C:
			log.Println("ping")
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Println("write message error: ", err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	log.Println("pong")
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *Client) Send(ev event.Event) {
	c.egress <- ev
}

func (c *Client) Close() {
	close(c.egress)
	c.conn.Close()
}
