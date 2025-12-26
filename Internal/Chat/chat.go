package chat

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ---- Config  ----
const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 5120
)

// ---- Utilities ----
func newID(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ---- Message format ----
type Message struct {
	Type    string `json:"type"`    // "message" or "system"
	Sender  string `json:"sender"`  // username
	Content string `json:"content"` // text
	Time    int64  `json:"time"`    // unix ms
}

// ---- Forward: HubAPI ----
// HubAPI is a minimal interface to allow chat.Room to call back to hub without import cycles.
type HubAPI interface {
	DeleteRoom(id string)
}

// ---- Room ----
type Room struct {
	ID         string
	clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
	closed     chan struct{}
	mu         sync.RWMutex
	timer      *time.Timer
	hub        HubAPI
}

func NewRoom(h HubAPI, lifetime time.Duration) *Room {
	r := &Room{
		ID:         newID(6),
		clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte, 256),
		closed:     make(chan struct{}),
		hub:        h,
	}
	// schedule close
	r.timer = time.AfterFunc(lifetime, func() {
		r.CloseRoom()
	})
	go r.run()
	return r
}

func (r *Room) run() {
	for {
		select {
		case c := <-r.Register:
			r.mu.Lock()
			r.clients[c] = true
			r.mu.Unlock()
			// system join message
			sys := Message{Type: "system", Sender: "system", Content: c.name + " joined", Time: time.Now().UnixMilli()}
			b, _ := json.Marshal(sys)
			r.Broadcast <- b

		case c := <-r.Unregister:
			r.mu.Lock()
			if _, ok := r.clients[c]; ok {
				delete(r.clients, c)
				close(c.send)
				// non-blocking system left
				sys := Message{Type: "system", Sender: "system", Content: c.name + " left", Time: time.Now().UnixMilli()}
				b, _ := json.Marshal(sys)
				select {
				case r.Broadcast <- b:
				default:
				}
			}
			r.mu.Unlock()

		case msg := <-r.Broadcast:
			r.mu.RLock()
			for c := range r.clients {
				// non-blocking send to avoid slow clients blocking room
				select {
				case c.send <- msg:
				default:
					// slow client => remove
					delete(r.clients, c)
					close(c.send)
				}
			}
			r.mu.RUnlock()

		case <-r.closed:
			// notify clients and close
			r.mu.RLock()
			for c := range r.clients {
				sys := Message{Type: "system", Sender: "system", Content: "room closed", Time: time.Now().UnixMilli()}
				b, _ := json.Marshal(sys)
				// best-effort write
				_ = c.conn.WriteMessage(websocket.TextMessage, b)
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				close(c.send)
				_ = c.conn.Close()
				delete(r.clients, c)
			}
			r.mu.RUnlock()
			// remove from hub
			if r.hub != nil {
				r.hub.DeleteRoom(r.ID)
			}
			return
		}
	}
}

func (r *Room) CloseRoom() {
	select {
	case <-r.closed:
		// already closed
	default:
		close(r.closed)
	}
}

// ---- Client ----
type Client struct {
	conn *websocket.Conn
	send chan []byte
	room *Room
	name string
}

func NewClient(conn *websocket.Conn, room *Room, name string) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte, 256),
		room: room,
		name: name,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		// unregister and close connection
		c.room.Unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMsgSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			// log.Println("read error:", err)
			break
		}
		// wrap plain text into Message
		m := Message{
			Type:    "message",
			Sender:  c.name,
			Content: string(msg),
			Time:    time.Now().UnixMilli(),
		}
		b, _ := json.Marshal(m)
		// send to room
		c.room.Broadcast <- b
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// channel closed by room
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(msg)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
