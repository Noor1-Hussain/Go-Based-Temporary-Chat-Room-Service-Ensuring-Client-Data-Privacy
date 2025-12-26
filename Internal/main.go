package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"code/internal/hub"
	ws "code/internal/websocket"
)

const (
	roomLifetime = 5 * time.Minute
)

func main() {
	h := hub.NewHub()

	// create-room endpoint (GET for simplicity)
	http.HandleFunc("/create-room", func(w http.ResponseWriter, r *http.Request) {
		room := h.CreateRoom(roomLifetime)
		resp := map[string]any{
			"room_id":          room.ID,
			"lifetime_minutes": int(roomLifetime.Minutes()),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	// WebSocket endpoint: /ws?room={id}&name={username}
	http.HandleFunc("/ws", ws.WSHandler(h))

	// health
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Chat backend running. POST /create-room -> {room_id}. Connect WS at /ws?room={id}&name={you}"))
	})

	addr := ":8080"
	log.Printf("listening on %s (room lifetime %v)\n", addr, roomLifetime)
	log.Fatal(http.ListenAndServe(addr, nil))
}
