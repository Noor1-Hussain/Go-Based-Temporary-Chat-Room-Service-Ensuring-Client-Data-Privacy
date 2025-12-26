package websocket

import (
	"log"
	"net/http"

	"code/internal/chat"
	"code/internal/hub"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// للـdev نسمح بأي Origin. في production لازم تقيديها.
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WSHandler(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Query().Get("room")
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "anon"
		}
		if roomID == "" {
			http.Error(w, "missing room param", http.StatusBadRequest)
			return
		}
		room, ok := h.GetRoom(roomID)
		if !ok {
			http.Error(w, "room not found or closed", http.StatusNotFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}

		client := chat.NewClient(conn, room, name)
		// register to room
		room.Register <- client

		// start pumps
		go client.WritePump()
		go client.ReadPump()
	}
}
