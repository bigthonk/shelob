package localstorage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Document represents a crawled page.
type Document struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Storage manages saving documents to disk.
type Storage struct {
	dir string
}

// NewStorage creates the storage directory (if needed) and returns a Storage instance.
func NewStorage(dir string) (*Storage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("could not create directory %s: %v", dir, err)
	}
	return &Storage{dir: dir}, nil
}

// SaveDocument writes the document to a JSON file in the storage directory.
// It creates a filename based on the URL.
func (s *Storage) SaveDocument(doc Document) error {
	// Create a safe filename by replacing "/" and ":" with underscores.
	fname := strings.ReplaceAll(doc.URL, "/", "_")
	fname = strings.ReplaceAll(fname, ":", "_")
	path := filepath.Join(s.dir, fname+".json")

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling document: %v", err)
	}
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing file %s: %v", path, err)
	}
	return nil
}
