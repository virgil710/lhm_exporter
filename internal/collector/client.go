package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout        = 10 * time.Second
	maxResponseSize       = 10 * 1024 * 1024 // 10 MB
	maxIdleConns          = 10
	maxIdleConnsPerHost   = 5
	defaultIdleConnTimeout = 90 * time.Second
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
		timeout = defaultTimeout
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = maxIdleConns
	transport.MaxIdleConnsPerHost = maxIdleConnsPerHost
	transport.IdleConnTimeout = defaultIdleConnTimeout

	return &LHMClient{
		url: url,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

// URL returns the target endpoint URL being accessed.
func (c *LHMClient) URL() string {
	return c.url
}

// Close releases any resources held by the client.
func (c *LHMClient) Close() {
	c.httpClient.CloseIdleConnections()
}

// Fetch retrieves and decodes the LHM data from the remote endpoint.
// It uses streaming JSON decoding with a size limit to prevent memory exhaustion.
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
	limitedBody := io.LimitReader(resp.Body, maxResponseSize)
	if err := json.NewDecoder(limitedBody).Decode(&node); err != nil {
		return nil, fmt.Errorf("decoding LHM data: %w", err)
	}

	return &node, nil
}
