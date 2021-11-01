package gosse

import (
	"net/http"
	"strconv"
)

const (
	MIMETextEventStream = "text/event-stream"
)

func (b *broker) StdHTTPHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Topic unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", MIMETextEventStream)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	eventSourceID := r.URL.Query().Get("topic")
	if eventSourceID == "" {
		http.Error(w, "Please specify a topic", http.StatusBadRequest)
		return
	}

	topic, err := b.GetTopic(eventSourceID)
	if err != nil {
		if err == ErrTopicNotExist && b.AutoCreation() {
			topic = b.CreateTopic(eventSourceID)
		} else if err == ErrTopicNotExist && !b.AutoCreation() {
			http.Error(w, "Topic not found", http.StatusNotFound)
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

	conn := topic.addConnection(eventID)

	go func() {
		<-r.Context().Done()
		conn.Close()
	}()

	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for event := range conn.Queue() {
		if len(event.Data()) == 0 {
			break
		}

		_, _ = w.Write(event.Bytes())

		flusher.Flush()
	}
}
