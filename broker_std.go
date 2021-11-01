package gosse

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	MIMETextEventStream = "text/event-stream"
)

func (b *broker) StdHTTPHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Event Source unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", MIMETextEventStream)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	eventSourceID := r.URL.Query().Get("source")
	if eventSourceID == "" {
		http.Error(w, "Please specify a event source", http.StatusBadRequest)
		return
	}

	source, err := b.GetEventSource(eventSourceID)
	if err != nil {
		if err == ErrEventSourceNotExist && b.AutoCreation() {
			source = b.CreateEventSource(eventSourceID)
		} else if err == ErrEventSourceNotExist && !b.AutoCreation() {
			http.Error(w, "Event Source not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	eventID := 0
	if id := r.Header.Get("Last-Event-ID"); id != "" {
		var err error
		eventID, err = strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Last-Event-ID must be a valid number format", http.StatusBadRequest)
			return
		}
	}

	conn := source.addConnection(eventID)

	go func() {
		<-r.Context().Done()
		conn.Close()
	}()

	flusher.Flush()

	for event := range conn.Queue() {
		if len(event.Data()) == 0 {
			break
		}

		_, _ = fmt.Fprintf(w, "id: %s\n", event.ID())
		_, _ = fmt.Fprintf(w, "data: %s\n", event.Data())
		_, _ = fmt.Fprintf(w, "event: %s\n", event.Event())
		_, _ = fmt.Fprintf(w, "retry: %s\n", event.Retry())
		_, _ = fmt.Fprint(w, "\n")

		flusher.Flush()
	}
}
