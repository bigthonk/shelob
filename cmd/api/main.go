package main

import (
	"log"

	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"search/api"
	"search/indexer"
	"search/localstorage"
)

// loadDocuments reads all JSON files from the given directory and indexes them.
func loadDocuments(dir string, idx *indexer.InvertedIndex) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(dir, file.Name())
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Error reading file %s: %v", path, err)
			continue
		}
		var doc localstorage.Document
		if err := json.Unmarshal(data, &doc); err != nil {
			log.Printf("Error unmarshaling file %s: %v", path, err)
			continue
		}
		idx.IndexDocument(indexer.Document{
			URL:   doc.URL,
			Title: doc.Title,
			Body:  doc.Body,
		})
	}
	return nil
}

func main() {
	// Create an in‑memory index.
	idx := indexer.NewInvertedIndex()

	// (Optional) Load previously saved documents from local storage.
	dataDir := "data"
	if err := loadDocuments(dataDir, idx); err != nil {
		log.Printf("Error loading documents: %v", err)
	} else {
		log.Println("Loaded documents from local storage into the index.")
	}

	// Start the API server.
	api.StartServer(idx)
}
