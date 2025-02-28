package main

import (
	"html/template"
	"log"
	"net/http"
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

func main() {
	http.HandleFunc("/", serveRoot)

	auth()

	log.Printf("Auth token: %v", storedToken.Token)

	addr := "localhost:9090"

	log.Printf("Serving on %s\n", addr)
	err := http.ListenAndServe("localhost:9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
