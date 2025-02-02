# Shelob

This project is a simple search engine written in Go. It includes:
- A **crawler** that fetches web pages, respects `robots.txt`, extracts text and links.
- An **indexer** that builds an inverted index of crawled pages.
- An **API** that exposes a search endpoint to query the index.

## Running the Project

1. Build and run the project:
   ```bash
  go run cmd/crawler/main.go
  go run cmd/api/main.go
  curl "http://localhost:8080/search?q=Dogs"



