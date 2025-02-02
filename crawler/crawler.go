package crawler

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"search/indexer"
	"search/localstorage"

	"golang.org/x/net/html"
)

// Run starts a simple single-threaded crawler that processes every enqueued URL.
// It uses a slice as a FIFO queue and writes each fetched page to disk.
func Run(seedURLs []string, store *localstorage.Storage, idx *indexer.InvertedIndex) {
	// Our work queue is just a slice of URLs.
	queue := make([]string, 0, len(seedURLs))
	queue = append(queue, seedURLs...)

	// Process URLs until the queue is empty.
	for len(queue) > 0 {
		// Dequeue the first URL.
		currentURL := queue[0]
		queue = queue[1:]
		log.Println("Processing:", currentURL)

		// Normalize the URL.
		parsed, err := url.Parse(currentURL)
		if err != nil {
			log.Println("Invalid URL:", currentURL, "error:", err)
			continue
		}
		parsed.Fragment = ""
		normalized := parsed.String()

		// Fetch the URL.
		resp, err := http.Get(normalized)
		if err != nil {
			log.Println("Error fetching", normalized, ":", err)
			continue
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Println("Error reading body for", normalized, ":", err)
			continue
		}
		log.Printf("Fetched %s (length=%d)", normalized, len(bodyBytes))

		// Parse HTML.
		doc, err := html.Parse(strings.NewReader(string(bodyBytes)))
		if err != nil {
			log.Println("Error parsing HTML from", normalized, ":", err)
			continue
		}

		// Extract title and text.
		title := extractTitle(doc)
		text := extractText(doc)

		// Index the document.
		idx.IndexDocument(indexer.Document{
			URL:   normalized,
			Title: title,
			Body:  text,
		})

		// Save the document to disk.
		err = store.SaveDocument(localstorage.Document{
			URL:   normalized,
			Title: title,
			Body:  text,
		})
		if err != nil {
			log.Println("Error saving document", normalized, ":", err)
		} else {
			log.Println("Saved document:", normalized)
		}

		// Extract links.
		links := extractLinks(doc)
		log.Printf("Found %d links on %s", len(links), normalized)
		// For each discovered link, resolve it and enqueue.
		for _, link := range links {
			abs := resolveURL(normalized, link)
			if abs != "" {
				queue = append(queue, abs)
				log.Println("Enqueued:", abs)
			}
		}

		// (Optional) For demonstration, you might limit the total number of processed pages.
		// if some counter exceeds a threshold, break out of the loop.
	}
	log.Println("Queue exhausted; crawler finished.")
}

// resolveURL resolves a possibly relative URL (href) against the base URL.
func resolveURL(baseStr, href string) string {
	base, err := url.Parse(baseStr)
	if err != nil {
		return ""
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return base.ResolveReference(ref).String()
}

// extractLinks recursively extracts all href attribute values from <a> tags.
func extractLinks(n *html.Node) []string {
	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					links = append(links, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return links
}

// extractText recursively extracts text content from the HTML node.
func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var result string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result += extractText(c)
	}
	return result
}

// extractTitle recursively finds the <title> element and returns its text.
func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
		return n.FirstChild.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		title := extractTitle(c)
		if title != "" {
			return title
		}
	}
	return ""
}
