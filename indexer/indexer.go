package indexer

import (
	"log"
	"strings"
	"sync"
)

// Document represents a crawled web page.
type Document struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// InvertedIndex is a simple in‑memory index that stores documents.
type InvertedIndex struct {
	documents []Document
	mu        sync.RWMutex
}

// NewInvertedIndex creates a new, empty index.
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		documents: make([]Document, 0),
	}
}

// IndexDocument adds a document to the index.
func (idx *InvertedIndex) IndexDocument(doc Document) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.documents = append(idx.documents, doc)
	log.Printf("Indexed document: %s (Title: %s)", doc.URL, doc.Title)
}

// Search returns a list of documents where the query appears
// in the title or the body (case-insensitive substring search).
func (idx *InvertedIndex) Search(query string) []Document {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	var results []Document
	query = strings.ToLower(query)
	for _, doc := range idx.documents {
		if strings.Contains(strings.ToLower(doc.Title), query) ||
			strings.Contains(strings.ToLower(doc.Body), query) {
			results = append(results, doc)
		}
	}
	return results
}
