package manager

import (
	"Likain/internal/client"
	"Likain/internal/event"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Manager struct {
	clients  client.ClientList
	handlers map[string]event.Handler
	sync.RWMutex
}

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}
)

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:8080":
		return true
	default:
		return false
	}
}

func NewManager() *Manager {
	m := &Manager{
		clients:  make(client.ClientList),
		handlers: make(map[string]event.Handler),
	}

	m.registerHandler()
	return m
}

func (m *Manager) registerHandler() {
	m.handlers[event.EventSendMessage] = event.HandleSendMessage
}

func (m *Manager) addClient(c *client.Client) {
	m.Lock()
	m.clients[c] = true
	m.Unlock()
}

func (m *Manager) RouteEvent(ev event.Event, c *client.Client) error {
	if handler, ok := m.handlers[ev.Type]; ok {
		return handler(ev, c)
	}
	return errors.New("unknown event type")
}

func (m *Manager) RemoveClient(c *client.Client) {
	m.Lock()
	if _, ok := m.clients[c]; ok {
		c.Close()
		delete(m.clients, c)
	}
	m.Unlock()
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c := client.New(conn, m)
	m.addClient(c)

	go c.ReadMessages()
	go c.WriteMessage()
}

func (m *Manager) broadcastMessage(ev event.Event) {
	m.RLock()
	defer m.RUnlock()

	for c := range m.clients {
		c.Send(ev)
	}
}
