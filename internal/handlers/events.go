// internal/handlers/events.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/nickg76/garage-backend/internal/auth"
)

type Event struct {
	Type 			string `json:"type"`
	Appointment     string `json:"appointment_id,omitempty"`
	Status          string `json:"status,omitempty"`
	Message         string `json:"message,omitempty"`
}

type EventHub struct {
	mu 			sync.RWMutex
	subs        map[string]map[chan []byte]struct{}
}

func NewEventHub() *EventHub {
	return &EventHub{
		subs: make(map[string]map[chan []byte]struct{}),
	}
}

func (h *EventHub) Subscribe(userID string) (chan []byte, func()) {
	ch := make(chan []byte, 8)
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.subs[userID]; !ok {
		h.subs[userID] = make(map[chan []byte]struct{})
	}
	h.subs[userID][ch] = struct{}{}
	unsub := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if set, ok := h.subs[userID]; ok {
			delete(set, ch)
			if len(set) == 0 {
				delete(h.subs, userID)
			}
		}
		close(ch)
	}
	return ch, unsub
}

func (h *EventHub) Publish(userID string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[userID] {
		select {
		case ch <-payload:
		default:
		}
	}
}

// SSE endpoint. Pass JWT via query param token because EventSource can't set custom headers.
func (s *Server) Events(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	claims, err := auth.ParseJWT(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	userID := claims.Sub
	if userID == "" {
		http.Error(w, "invalid user", http.StatusUnauthorized)
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}

	ch, unsubscribe := s.hub.Subscribe(userID)
	defer unsubscribe()

	// initial comment to open the stream
	_, _ = w.Write([]byte(": connected\n\n"))
	flusher.Flush()

	// heatbeat to keep connections alive (proxies)
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = w.Write([]byte(": ping\n\n"))
			flusher.Flush()
		case msg, ok := <-ch:
			if !ok {
				return
			}
			// write SSE data 
			w.Write([]byte("data: "))
			w.Write(msg)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}

// Helper to publish typed events
func (s *Server) publishToUser(ctx context.Context, userID string, ev Event) {
	b, _ := json.Marshal(ev)
	s.hub.Publish(userID, b)
}
