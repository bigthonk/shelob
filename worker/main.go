package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html"

	"shelob/worker/robots"
)

// PageData holds information about the fetched page.
type PageData struct {
	URL   string
	Title string
}

func main() {
	frontierURL := os.Getenv("FRONTIER_URL")
	if frontierURL == "" {
		frontierURL = "http://frontier:8080"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down worker...")
		cancel()
	}()

	httpClient := &http.Client{Timeout: 10 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			urls := fetchBatch(frontierURL, 10)
			if len(urls) == 0 {
				// Nothing to do, wait before checking again
				time.Sleep(3 * time.Second)
				continue
			}

			for _, u := range urls {
				allowed, err := robots.CheckAllowed(httpClient, u)
				if err != nil {
					log.Printf("Error checking robots for %s: %v", u, err)
					continue
				}
				if !allowed {
					log.Printf("Skipping %s due to robots.txt disallow rules", u)
					continue
				}

				pageData, links, err := fetchAndParseReal(httpClient, u)
				if err != nil {
					log.Printf("Error processing %s: %v", u, err)
					continue
				}
				log.Printf("Processed page: %s with title: %s", pageData.URL, pageData.Title)

				if len(links) > 0 {
					err = addURLs(frontierURL, links)
					if err != nil {
						log.Printf("Failed to add URLs: %v", err)
					}
				}
			}
		}
	}
}

// fetchBatch retrieves a batch of URLs from the frontier.
func fetchBatch(frontierURL string, batchSize int) []string {
	endpoint := fmt.Sprintf("%s/fetch?batch=%d", frontierURL, batchSize)
	resp, err := http.Get(endpoint)
	if err != nil {
		log.Printf("Error fetching batch from frontier: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Frontier returned non-OK status: %d", resp.StatusCode)
		return nil
	}

	var urls []string
	err = json.NewDecoder(resp.Body).Decode(&urls)
	if err != nil {
		log.Printf("Error decoding frontier response: %v", err)
		return nil
	}

	return urls
}

// addURLs sends newly discovered URLs to the frontier.
func addURLs(frontierURL string, urls []string) error {
	data := map[string]interface{}{
		"urls": urls,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling urls: %w", err)
	}

	resp, err := http.Post(frontierURL+"/add", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error adding urls to frontier: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("frontier returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}

// fetchAndParseReal fetches a page, parses it to extract the title and links.
func fetchAndParseReal(httpClient *http.Client, urlStr string) (PageData, []string, error) {
	resp, err := httpClient.Get(urlStr)
	if err != nil {
		return PageData{}, nil, fmt.Errorf("error fetching page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PageData{}, nil, fmt.Errorf("non-OK HTTP status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return PageData{}, nil, fmt.Errorf("error reading page body: %w", err)
	}

	pageData, links, err := parseHTML(data, urlStr)
	if err != nil {
		return PageData{}, nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	return pageData, links, nil
}

// parseHTML parses the given HTML data, extracts the page title and links.
func parseHTML(data []byte, baseURL string) (PageData, []string, error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return PageData{}, nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	var links []string
	var title string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "title" && n.FirstChild != nil {
				title = strings.TrimSpace(n.FirstChild.Data)
			}
			if n.Data == "a" {
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						u, err := resolveURL(baseURL, attr.Val)
						if err == nil {
							links = append(links, u)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return PageData{URL: baseURL, Title: title}, links, nil
}

// resolveURL resolves a possibly relative URL against a base URL.
func resolveURL(baseURL, href string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base url: %w", err)
	}
	u, err := url.Parse(strings.TrimSpace(href))
	if err != nil {
		return "", fmt.Errorf("invalid href: %w", err)
	}
	resolved := base.ResolveReference(u)
	if resolved.Scheme == "" || resolved.Host == "" {
		return "", errors.New("resolved URL is missing scheme or host")
	}
	return resolved.String(), nil
}
