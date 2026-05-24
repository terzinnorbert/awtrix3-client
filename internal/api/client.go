package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the AWTRIX 3 HTTP API client.
type Client struct {
	host       string
	httpClient *http.Client
}

// NewClient creates a new Client for the given host (IP or hostname, no scheme).
func NewClient(host string) *Client {
	return &Client{
		host: host,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Host returns the configured host.
func (c *Client) Host() string { return c.host }

func (c *Client) baseURL() string {
	return fmt.Sprintf("http://%s/api", c.host)
}

// get performs a GET request and JSON-decodes the response into out.
func (c *Client) get(path string, out interface{}) error {
	resp, err := c.httpClient.Get(c.baseURL() + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// post performs a POST request, JSON-encoding payload as the request body.
// A nil payload sends an empty body (used for "clear" / "dismiss" endpoints).
func (c *Client) post(path string, payload interface{}) error {
	var body io.Reader = bytes.NewReader([]byte{})
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	resp, err := c.httpClient.Post(c.baseURL()+path, "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// postRaw sends a plain-text POST (used for the /rtttl endpoint).
func (c *Client) postRaw(path string, payload string) error {
	resp, err := c.httpClient.Post(c.baseURL()+path, "text/plain", strings.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}
