package gosse

import "strconv"

type EventWAL []Event

func (wal *EventWAL) Add(e Event) {
	e.(*event).id = wal.currentIdx()
	*wal = append(*wal, e)
}

func (wal *EventWAL) Reply(c Connection) {
	for idx := range *wal {
		id, _ := strconv.Atoi((*wal)[idx].ID())
		if id >= c.EventID() {
			c.Send((*wal)[idx])
		}
	}
}

func (wal *EventWAL) Purge() {
	*wal = nil
}

func (wal *EventWAL) currentIdx() string {
	return strconv.Itoa(len(*wal))
}
