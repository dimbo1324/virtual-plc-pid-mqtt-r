package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ssePayload struct {
	name string
	data []byte
}

type sseBroker struct {
	mu      sync.Mutex
	clients map[chan ssePayload]struct{}
}

func newSSEBroker() *sseBroker {
	return &sseBroker{clients: make(map[chan ssePayload]struct{})}
}

func (b *sseBroker) subscribe() chan ssePayload {
	ch := make(chan ssePayload, 16)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *sseBroker) unsubscribe(ch chan ssePayload) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
}

func (b *sseBroker) clientCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.clients)
}

// broadcast serializes v as JSON and sends it to all subscribed clients.
// Slow clients that would block are silently skipped.
func (b *sseBroker) broadcast(name string, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	p := ssePayload{name: name, data: data}
	b.mu.Lock()
	for ch := range b.clients {
		select {
		case ch <- p:
		default:
		}
	}
	b.mu.Unlock()
}

func (b *sseBroker) serveSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush() // send headers to the client immediately

	ch := b.subscribe()
	defer b.unsubscribe(ch)

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case p := <-ch:
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", p.name, p.data)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprintf(w, "event: heartbeat\ndata: {}\n\n")
			flusher.Flush()
		}
	}
}
