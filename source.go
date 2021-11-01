package gosse

type Topic struct {
	id     string
	replay bool

	conn []Connection
	wal  EventWAL

	quitC      chan struct{}
	events     chan Event
	register   chan Connection
	deregister chan Connection
}

func newTopic(id string, bufferSize int, playable bool) *Topic {
	return &Topic{
		id:         id,
		replay:     playable,
		conn:       make([]Connection, 0),
		wal:        make(EventWAL, 0),
		quitC:      make(chan struct{}),
		events:     make(chan Event, bufferSize),
		register:   make(chan Connection),
		deregister: make(chan Connection),
	}
}

func (src *Topic) addConnection(eventID int) Connection {
	conn := &conn{
		eventID: eventID,
		quit:    make(chan struct{}),
		queue:   make(chan Event, 64),
	}

	src.register <- conn
	return conn
}

func (src *Topic) getConnectionIndex(conn Connection) int {
	for idx, connection := range src.conn {
		if connection == conn {
			return idx
		}
	}
	return -1
}

func (src *Topic) removeConnection(idx int) {
	src.conn[idx].Close()
	src.conn = append(src.conn[:idx], src.conn[idx+1:]...)
}

func (src *Topic) removeAllConnections() {
	for _, connection := range src.conn {
		connection.Close()
	}
	src.conn = src.conn[:0]
}

func (src *Topic) close() {
	src.quitC <- struct{}{}
}

func (src *Topic) background() {
BackgroundLoop:
	for {
		select {
		// new subscriber coming
		case sub := <-src.register:
			src.conn = append(src.conn, sub)
			if src.replay {
				src.wal.Reply(sub)
			}
		// remove subscriber
		case sub := <-src.deregister:
			idx := src.getConnectionIndex(sub)
			if idx != -1 {
				src.removeConnection(idx)
			}
		// publish event
		case event := <-src.events:
			if src.replay {
				src.wal.Add(event)
			}
			for _, conn := range src.conn {
				conn.Send(event)
			}
		// shutdown
		case <-src.quitC:
			src.removeAllConnections()
			break BackgroundLoop
		}
	}
}
