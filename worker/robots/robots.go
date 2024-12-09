package robots

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// RobotsRules holds the Disallow rules for a given domain.
type RobotsRules struct {
	Disallow []string
}

var robotsCache = struct {
	mu     sync.RWMutex
	domain map[string]*RobotsRules
}{
	domain: make(map[string]*RobotsRules),
}

// CheckAllowed checks if a given URL is allowed by the domain's robots.txt rules.
func CheckAllowed(client *http.Client, rawURL string) (bool, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false, err
	}

	domain := u.Host
	rules := getRobotsRules(domain)
	if rules == nil {
		// Not in cache, fetch
		rules, err = fetchAndParseRobots(client, u.Scheme, domain)
		if err != nil {
			// If we can’t fetch robots.txt, we can choose to allow by default:
			log.Printf("Could not fetch robots for %s, allowing by default. Error: %v", domain, err)
			return true, nil
		}
		storeRobotsRules(domain, rules)
	}

	path := u.EscapedPath()
	for _, d := range rules.Disallow {
		if strings.HasPrefix(path, d) {
			return false, nil
		}
	}
	return true, nil
}

func fetchAndParseRobots(client *http.Client, scheme, domain string) (*RobotsRules, error) {
	robotsURL := scheme + "://" + domain + "/robots.txt"
	resp, err := client.Get(robotsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No robots.txt means no restrictions
		return &RobotsRules{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status %d fetching robots.txt from %s", resp.StatusCode, domain)
	}

	return parseRobots(resp.Body), nil
}

func parseRobots(r io.Reader) *RobotsRules {
	scanner := bufio.NewScanner(r)
	var inDefaultAgent bool
	var disallows []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lc := strings.ToLower(line)
		if strings.HasPrefix(lc, "user-agent:") {
			ua := strings.TrimSpace(line[len("User-agent:"):])
			ua = strings.ToLower(ua)
			inDefaultAgent = (ua == "*")
		} else if inDefaultAgent && strings.HasPrefix(lc, "disallow:") {
			path := strings.TrimSpace(line[len("Disallow:"):])
			if path != "" {
				disallows = append(disallows, path)
			}
		}
	}

	return &RobotsRules{Disallow: disallows}
}

func storeRobotsRules(domain string, rules *RobotsRules) {
	robotsCache.mu.Lock()
	defer robotsCache.mu.Unlock()
	robotsCache.domain[domain] = rules
}

func getRobotsRules(domain string) *RobotsRules {
	robotsCache.mu.RLock()
	defer robotsCache.mu.RUnlock()
	return robotsCache.domain[domain]
}
