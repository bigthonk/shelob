package api

import (
	"encoding/json"
	"log"
	"net/http"
	"search/indexer"
)

// searchHandler returns an HTTP handler function that performs a search on the index.
func searchHandler(idx *indexer.InvertedIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
			return
		}
		results := idx.Search(query)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(results); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// StartServer starts the HTTP API server on port 8080.
func StartServer(idx *indexer.InvertedIndex) {
	http.HandleFunc("/search", searchHandler(idx))
	log.Println("Search API listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
