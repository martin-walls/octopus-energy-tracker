package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

func serveRoot(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	if r.URL.Path != "/" {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmplt, err := template.ParseFiles("home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// query := struct {
	// 	Query string `json:"query"`
	// }{
	// 	Query: `
	// 		query GreennessForecast {
	// 			greennessForecast {
	// 				validFrom
	// 				validTo
	// 				greennessScore
	// 				greennessIndex
	// 				highlightFlag
	// 			}
	// 		}
	// 	`,
	// }
	//
	// body, err := json.Marshal(query)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	//
	// request, err := http.NewRequest(http.MethodPost, "https://api.octopus.energy/v1/graphql/", bytes.NewBuffer(body))
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// request.Header.Add("Content-Type", "application/json")
	//
	// client := &http.Client{}
	//
	// response, err := client.Do(request)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// defer response.Body.Close()
	//
	// responseBytes, err := io.ReadAll(response.Body)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	//
	// responseBody := string(responseBytes)

	tmplt.Execute(w, struct {
		CurrentWatts int
		Greenness    string
	}{
		CurrentWatts: 5,
		Greenness:    "",
	})
}

type WebsocketHandler struct {
	broadcaster *Broadcaster[*ConsumptionReading]
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

func pollLiveConsumption(b *Broadcaster[*ConsumptionReading]) {
	octopus := Octopus{}

	for {
		reading, err := octopus.LiveConsumption()
		if err != nil {
			if errors.Is(err, ErrSkippingRequest) || errors.Is(err, ErrTooManyRequests) {
				log.Println(err)
			} else {
				log.Fatalln(err)
			}
		} else {
			log.Printf("Using %vW", reading.Demand)
			b.Publish(reading)
		}

		time.Sleep(10 * time.Second)
	}
}

func main() {
	b := NewBroadcaster[*ConsumptionReading]()

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
