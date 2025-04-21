package main

import (
	"encoding/json"
	"errors"
	"log"
	"martin-walls/octopus-energy-tracker/internal/broadcaster"
	"martin-walls/octopus-energy-tracker/internal/octopus"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

type WebsocketHandler struct {
	broadcaster *broadcaster.Broadcaster[*octopus.ConsumptionReading]
}

func (wsHandler *WebsocketHandler) handle(w http.ResponseWriter, r *http.Request) {
	log.Println("Got websocket connection")
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("Failed to accept websocket connection: %v", err)
		return
	}
	defer c.CloseNow()

	// Allow websocket to be open for at most one minute
	// Handle this as a write-only websocket
	ctx := c.CloseRead(r.Context())

	readings := wsHandler.broadcaster.Subscribe()

	for {
		select {
		case <-ctx.Done():
			c.Close(websocket.StatusNormalClosure, "")
			log.Println("Closing websocket")
			return
		case reading := <-readings:
			readingJson, err := json.Marshal(reading)
			if err != nil {
				log.Printf("Failed to JSON encode consumption reading: %v", reading)
				continue
			}

			err = c.Write(ctx, websocket.MessageText, readingJson)

			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				log.Println("Closing websocket")
				return
			} else if err != nil {
				log.Printf("err: %v", err)
				return
			}
		}
	}
}

func pollLiveConsumption(b *broadcaster.Broadcaster[*octopus.ConsumptionReading]) {
	octo := octopus.Octopus{}

	for {
		reading, err := octo.LiveConsumption()
		if err != nil {
			if errors.Is(err, octopus.ErrSkippingRequest) || errors.Is(err, octopus.ErrTooManyRequests) {
				log.Println(err)
			} else {
				log.Fatalln(err)
			}
		} else {
			log.Printf("Using %vW", reading.Demand)
			b.Publish(reading)
		}

		time.Sleep(30 * time.Second)
	}
}

func main() {
	b := broadcaster.NewBroadcaster[*octopus.ConsumptionReading]()

	go b.Start()
	defer b.Stop()

	go pollLiveConsumption(b)

	http.Handle("/", http.FileServer(http.Dir("static")))

	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("foo"))
	})

	wsHandler := WebsocketHandler{
		broadcaster: b,
	}
	http.HandleFunc("/ws", wsHandler.handle)

	addr := "localhost:9090"

	log.Printf("Serving on %s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
