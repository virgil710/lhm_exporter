package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LHMClient fetches hardware monitoring data from a LibreHardwareMonitor
// web server instance.
type LHMClient struct {
	httpClient *http.Client
	url        string
	fetchFn    func() (*Node, error)
}

// NewLHMClient creates a client for the given LHM endpoint.
func NewLHMClient(destIP string, destPort uint, timeout time.Duration) *LHMClient {
	url := fmt.Sprintf("http://%s:%d/data.json", destIP, destPort)
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &LHMClient{
		url: url,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Fetch retrieves and decodes the LHM data from the remote endpoint.
// It uses streaming JSON decoding to minimize memory allocation.
func (c *LHMClient) Fetch() (*Node, error) {
	if c.fetchFn != nil {
		return c.fetchFn()
	}

	resp, err := c.httpClient.Get(c.url)
	if err != nil {
		return nil, fmt.Errorf("fetching LHM data from %s: %w", c.url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LHM returned HTTP %d from %s", resp.StatusCode, c.url)
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("decoding LHM data: %w", err)
	}

	return &node, nil
}
