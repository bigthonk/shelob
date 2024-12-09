package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Frontier struct {
	mu   sync.Mutex
	urls []string
}

var frontier = &Frontier{}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/add", handleAdd).Methods("POST")
	r.HandleFunc("/fetch", handleFetch).Methods("GET")

	log.Println("Frontier running on :8080")
	http.ListenAndServe(":8080", r)
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URLs []string `json:"urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	frontier.mu.Lock()
	frontier.urls = append(frontier.urls, req.URLs...)
	frontier.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func handleFetch(w http.ResponseWriter, r *http.Request) {
	batchParam := r.URL.Query().Get("batch")
	if batchParam == "" {
		batchParam = "10"
	}
	batchSize, _ := strconv.Atoi(batchParam)

	frontier.mu.Lock()
	if len(frontier.urls) == 0 {
		frontier.mu.Unlock()
		json.NewEncoder(w).Encode([]string{})
		return
	}

	end := batchSize
	if end > len(frontier.urls) {
		end = len(frontier.urls)
	}
	fetched := frontier.urls[:end]
	frontier.urls = frontier.urls[end:]
	frontier.mu.Unlock()

	json.NewEncoder(w).Encode(fetched)
}
