package gosse

import (
	"bytes"
	"fmt"
	"strings"
)

// Event represents an event that occurs on the server
// and needs to be pushed to the client, and consists
// of Data and Event, with Data representing the payload
// of the event and Event representing the event type
type Event interface {
	ID() string

	Data() string

	Event() string

	Retry() int

	Buffer() *bytes.Buffer

	String() string

	Bytes() []byte
}

type event struct {
	id,
	data,
	event string
	retry int
}

func (e *event) ID() string {
	return e.id
}

func (e *event) Data() string {
	return e.data
}

func (e *event) Event() string {
	return e.event
}

func (e *event) Retry() int {
	return e.retry
}

func (e *event) Buffer() *bytes.Buffer {
	var buffer bytes.Buffer

	if len(e.id) > 0 {
		buffer.WriteString(fmt.Sprintf("id: %s\n", e.id))
	}

	if e.retry > 0 {
		buffer.WriteString(fmt.Sprintf("retry: %d\n", e.retry))
	}

	if len(e.event) > 0 {
		buffer.WriteString(fmt.Sprintf("event: %s\n", e.event))
	}

	if len(e.data) > 0 {
		buffer.WriteString(fmt.Sprintf("data: %s\n", strings.Replace(e.data, "\n", "\ndata: ", -1)))
	}

	buffer.WriteString("\n")

	return &buffer
}

func (e *event) String() string {
	return e.Buffer().String()
}

func (e *event) Bytes() []byte {
	return e.Buffer().Bytes()
}

var _ Event = new(event)

func NewEvent(id, data, eventName string) Event {
	return &event{
		id:    id,
		data:  data,
		event: eventName,
		retry: 0,
	}
}

func SimpleEvent(data string) Event {
	return NewEvent("", data, "")
}
