package hub

import (
	"sync"
	"time"

	"code/internal/chat"
)

type Hub struct {
	rooms map[string]*chat.Room
	mu    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*chat.Room),
	}
}

func (h *Hub) CreateRoom(lifetime time.Duration) *chat.Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	r := chat.NewRoom(h, lifetime)
	h.rooms[r.ID] = r
	return r
}

func (h *Hub) GetRoom(id string) (*chat.Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	r, ok := h.rooms[id]
	return r, ok
}

func (h *Hub) DeleteRoom(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms, id)
}
