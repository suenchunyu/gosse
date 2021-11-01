package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/suenchunyu/gosse"
)

func main() {
	broker := gosse.NewBroker(3, true, true)

	mux := http.NewServeMux()
	mux.HandleFunc("/events", broker.StdHTTPHandler)

	go generateMessage(broker)

	http.ListenAndServe(":7897", mux)
}

func generateMessage(broker gosse.Broker) {
	ticker := time.NewTicker(5 * time.Second)
	idx := 1
	for {
		select {
		case <-ticker.C:
			idxStr := strconv.Itoa(idx)

			if err := broker.Publish("test", gosse.NewEvent([]byte(idxStr), []byte("test"), []byte("test"), []byte("1"))); err != nil {
				panic(err)
			}
			idx++
			ticker.Reset(5 * time.Second)
		}
	}
}
