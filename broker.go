package gosse

import (
	"errors"
	"net/http"
	"sync"
)

var (
	ErrEventSourceNotExist = errors.New("specified event source not exist")
)

type Broker interface {
	Publish(id string, event Event) error

	CreateEventSource(id string) *EventSource

	DeleteEventSource(id string) error

	GetEventSource(id string) (*EventSource, error)

	StdHTTPHandler(w http.ResponseWriter, r *http.Request)

	AutoCreation() bool

	Close() error
}

type broker struct {
	mu *sync.Mutex

	bufferSize   int
	playable     bool
	autoCreation bool

	sources map[string]*EventSource
}

var _ Broker = new(broker)

func NewBroker(bufferSize int, playable bool, autoCreation bool) Broker {
	return &broker{
		mu:           new(sync.Mutex),
		bufferSize:   bufferSize,
		playable:     playable,
		autoCreation: autoCreation,
		sources:      make(map[string]*EventSource),
	}
}

func (b *broker) Publish(id string, event Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if source, exist := b.sources[id]; exist {
		source.events <- event
	} else {
		return ErrEventSourceNotExist
	}

	return nil
}

func (b *broker) CreateEventSource(id string) *EventSource {
	b.mu.Lock()
	defer b.mu.Unlock()

	if source, exist := b.sources[id]; exist {
		return source
	}

	src := newEventSource(id, b.bufferSize, b.playable)
	go src.background()

	b.sources[id] = src

	return src
}

func (b *broker) DeleteEventSource(id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if source, exist := b.sources[id]; exist {
		source.close()
		delete(b.sources, id)
	} else {
		return ErrEventSourceNotExist
	}

	return nil
}

func (b *broker) GetEventSource(id string) (*EventSource, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if source, exist := b.sources[id]; exist {
		return source, nil
	} else {
		return nil, ErrEventSourceNotExist
	}
}

func (b *broker) AutoCreation() bool {
	return b.autoCreation
}

func (b *broker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for id, source := range b.sources {
		source.close()
		delete(b.sources, id)
	}

	return nil
}
