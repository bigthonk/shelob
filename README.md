# Shelob - Simple Search Engine

Shelob is a minimalistic search engine built in Go. It consists of three core components:

- **Crawler:** Fetches web pages, respects `robots.txt`, and extracts text and links from crawled pages.
- **Indexer:** Builds an inverted index based on the crawled pages to facilitate fast search.
- **Search API:** Exposes a search endpoint for querying the index.

## Getting Started

Follow these steps to run the project locally:

1. **Start the Crawler**  
   The crawler will begin fetching and indexing web pages. Run the following command:
   ```bash
   go run cmd/crawler/main.go

2. **Start the Search API**
   The seach API will allow you to search through indexed pages. Run the following command:
   ```bash
   go run cmd/api/main.go
