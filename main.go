package main

import (
	"context"
	"fmt"
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

func serveWs(w http.ResponseWriter, r *http.Request) {
	log.Println("Got websocket connection")
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("Failed to accept websocket connection: %v", err)
		return
	}
	defer c.CloseNow()

	// Allow websocket to be open for at most one minute
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()

	// Handle this as a write-only websocket
	ctx = c.CloseRead(ctx)

	t := time.Tick(1 * time.Second)

	i := 0

	for {
		select {
		case <-ctx.Done():
			c.Close(websocket.StatusNormalClosure, "")
			log.Println("Closing websocket as timeout expired")
			return
		case <-t:
			err = c.Write(ctx, websocket.MessageText, fmt.Appendf(nil, "Value is %d", i))
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				log.Println("Closing websocket as timeout expired")
				return
			} else if err != nil {
				log.Printf("err: %v", err)
				return
			}
			i += 1
		}
	}
}

func main() {
	// octopus := Octopus{}
	//
	// for {
	// 	reading, err := octopus.LiveConsumption()
	// 	if err != nil {
	// 		if errors.Is(err, ErrSkippingRequest) || errors.Is(err, ErrTooManyRequests) {
	// 			log.Println(err)
	// 		} else {
	// 			log.Fatalln(err)
	// 		}
	// 	} else {
	// 		log.Printf("Using %vW", reading.Demand)
	// 	}
	//
	// 	time.Sleep(10 * time.Second)
	// }

	http.Handle("/", http.FileServer(http.Dir("static")))

	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("foo"))
	})

	http.HandleFunc("/ws", serveWs)

	addr := "localhost:9090"

	log.Printf("Serving on %s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
