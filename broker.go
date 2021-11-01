package gosse

import (
	"errors"
	"net/http"
	"sync"
)

var (
	ErrTopicNotExist = errors.New("specified topic not exist")
)

// Broker represents a proxy that the server applies
// to the Topic, providing programming interfaces
// such as creating Topic, publishing events.
type Broker interface {
	// Publish an event to the Topic with specified id
	Publish(id string, event Event) error

	// CreateTopic with specified id
	CreateTopic(id string) *Topic

	// DeleteTopic with specified id and release all connections
	DeleteTopic(id string) error

	// GetTopic with specified id
	GetTopic(id string) (*Topic, error)

	// StdHTTPHandler is the default handler based on net/http
	StdHTTPHandler(w http.ResponseWriter, r *http.Request)

	// AutoCreation returns true if creation of non-existent Topic
	// automatically
	AutoCreation() bool

	// Close broker and release all Topic and Connection.
	Close() error
}

type broker struct {
	mu *sync.Mutex

	bufferSize   int
	playable     bool
	autoCreation bool

	topics map[string]*Topic
}

var _ Broker = new(broker)

// NewBroker with specified bufferSize, could be set playable is true when
// you need replay event for newest connection, set autoCreation is true when
// creation of non-existent Topic automatically
func NewBroker(bufferSize int, playable bool, autoCreation bool) Broker {
	return &broker{
		mu:           new(sync.Mutex),
		bufferSize:   bufferSize,
		playable:     playable,
		autoCreation: autoCreation,
		topics:       make(map[string]*Topic),
	}
}

func (b *broker) Publish(id string, event Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if source, exist := b.topics[id]; exist {
		source.events <- event
	} else {
		return ErrTopicNotExist
	}

	return nil
}

func (b *broker) CreateTopic(id string) *Topic {
	b.mu.Lock()
	defer b.mu.Unlock()

	if source, exist := b.topics[id]; exist {
		return source
	}

	src := newTopic(id, b.bufferSize, b.playable)
	go src.background()

	b.topics[id] = src

	return src
}

func (b *broker) DeleteTopic(id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if source, exist := b.topics[id]; exist {
		source.close()
		delete(b.topics, id)
	} else {
		return ErrTopicNotExist
	}

	return nil
}

func (b *broker) GetTopic(id string) (*Topic, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if source, exist := b.topics[id]; exist {
		return source, nil
	} else {
		return nil, ErrTopicNotExist
	}
}

func (b *broker) AutoCreation() bool {
	return b.autoCreation
}

func (b *broker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for id, source := range b.topics {
		source.close()
		delete(b.topics, id)
	}

	return nil
}
