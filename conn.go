package gosse

type Connection interface {
	EventID() int
	Send(e Event)
	Close()
	Queue() <-chan Event
}

type conn struct {
	eventID int
	quit    chan struct{}
	queue   chan Event
}

func (c *conn) Queue() <-chan Event {
	return c.queue
}

var _ Connection = new(conn)

func (c *conn) EventID() int {
	return c.eventID
}

func (c *conn) Send(e Event) {
	c.queue <- e
}

func (c *conn) Close() {
	defer close(c.queue)
	c.quit <- struct{}{}
}
