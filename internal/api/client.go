package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client is the AWTRIX 3 HTTP API client.
// When multiple hosts are configured, write operations (POST) are broadcast
// concurrently to all hosts. Read operations (GET) target the first host only.
type Client struct {
	hosts      []string
	httpClient *http.Client
}

// NewClient creates a new Client for the given host (IP or hostname, no scheme).
func NewClient(host string) *Client {
	return &Client{
		hosts: []string{host},
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewMultiClient creates a Client that broadcasts write operations to all hosts.
// Read operations target the first host.
func NewMultiClient(hosts []string) *Client {
	return &Client{
		hosts: hosts,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Host returns the primary (first) configured host.
func (c *Client) Host() string {
	if len(c.hosts) > 0 {
		return c.hosts[0]
	}
	return ""
}

// Hosts returns all configured hosts.
func (c *Client) Hosts() []string { return c.hosts }

func (c *Client) baseURLFor(host string) string {
	return fmt.Sprintf("http://%s/api", host)
}

// get performs a GET request against the primary host and JSON-decodes the response into out.
func (c *Client) get(path string, out interface{}) error {
	resp, err := c.httpClient.Get(c.baseURLFor(c.Host()) + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// post broadcasts a JSON POST request concurrently to all configured hosts.
// A nil payload sends an empty body (used for "clear" / "dismiss" endpoints).
func (c *Client) post(path string, payload interface{}) error {
	var bodyBytes []byte
	if payload != nil {
		var err error
		bodyBytes, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	} else {
		bodyBytes = []byte{}
	}

	type hostErr struct {
		host string
		err  error
	}
	errs := make([]hostErr, len(c.hosts))
	var wg sync.WaitGroup
	for i, host := range c.hosts {
		wg.Add(1)
		go func(idx int, h string) {
			defer wg.Done()
			resp, err := c.httpClient.Post(
				c.baseURLFor(h)+path,
				"application/json",
				bytes.NewReader(bodyBytes),
			)
			if err != nil {
				errs[idx] = hostErr{h, err}
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				errs[idx] = hostErr{h, fmt.Errorf("HTTP %d", resp.StatusCode)}
			}
		}(i, host)
	}
	wg.Wait()

	var msgs []string
	for _, e := range errs {
		if e.err != nil {
			msgs = append(msgs, fmt.Sprintf("%s: %v", e.host, e.err))
		}
	}
	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "; "))
	}
	return nil
}

// postRaw broadcasts a plain-text POST concurrently to all configured hosts
// (used for the /rtttl endpoint).
func (c *Client) postRaw(path string, payload string) error {
	type hostErr struct {
		host string
		err  error
	}
	errs := make([]hostErr, len(c.hosts))
	var wg sync.WaitGroup
	for i, host := range c.hosts {
		wg.Add(1)
		go func(idx int, h string) {
			defer wg.Done()
			resp, err := c.httpClient.Post(
				c.baseURLFor(h)+path,
				"text/plain",
				strings.NewReader(payload),
			)
			if err != nil {
				errs[idx] = hostErr{h, err}
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				errs[idx] = hostErr{h, fmt.Errorf("HTTP %d", resp.StatusCode)}
			}
		}(i, host)
	}
	wg.Wait()

	var msgs []string
	for _, e := range errs {
		if e.err != nil {
			msgs = append(msgs, fmt.Sprintf("%s: %v", e.host, e.err))
		}
	}
	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "; "))
	}
	return nil
}
