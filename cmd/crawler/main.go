package main

import (
	"log"
	"time"

	"search/crawler"
	"search/indexer"
	"search/localstorage"
)

func main() {
	// Create the in‑memory index.
	idx := indexer.NewInvertedIndex()

	// Initialize local storage (documents will be saved as JSON files in the "data" directory).
	store, err := localstorage.NewStorage("data")
	if err != nil {
		log.Fatalf("Failed to initialize local storage: %v", err)
	}

	// Define seed URLs.
	seedURLs := []string{
		"https://en.wikipedia.org/wiki/Main_Page",
		"https://en.wikipedia.org/wiki/Portal:Current_events",
	}

	// Run the simple crawler in a separate goroutine.
	go crawler.Run(seedURLs, store, idx)

	// Keep the program running.
	for {
		time.Sleep(5 * time.Minute)
	}
}
