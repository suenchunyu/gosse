package gosse

type Event interface {
	ID() []byte
	SetID(id []byte)
	Data() []byte
	Event() []byte
	Retry() []byte
}

type event struct {
	id    []byte
	data  []byte
	event []byte
	retry []byte
}

var _ Event = new(event)

func NewEvent(id, data, eventName, retry []byte) Event {
	return &event{
		id:    id,
		data:  data,
		event: eventName,
		retry: retry,
	}
}

func (e event) ID() []byte {
	return e.id
}

func (e *event) SetID(id []byte) {
	e.id = id
}

func (e event) Data() []byte {
	return e.data
}

func (e event) Event() []byte {
	return e.event
}

func (e event) Retry() []byte {
	return e.retry
}
