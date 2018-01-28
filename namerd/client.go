package namerd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type Dtab string

func (d Dtab) String() string {
	return string(d)
}

// Dentries represents one or more Dentry
// Delegation tables are made up of one or mor Dentries
type Dentries []Dentry

// A Dentry reprents a single Delegation table rule
type Dentry struct {
	Prefix string `json:"prefix"`
	Dst    string `json:"dst"`
}

// String returns the Dentry in string format: /prefix => /destination;
func (d Dentry) String() string {
	return fmt.Sprintf("%s => %s;", d.Prefix, d.Dst)
}

// An Option function overrides Client configuration options
type Option func(*Client)

// WithClient overrides the configured http client
func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

// WithHost overrides the configured namerd host string
func WithHost(host string) Option {
	return func(c *Client) {
		c.host = host
	}
}

// WithScheme overrides the configured namerd request scheme
func WithScheme(scheme string) Option {
	return func(c *Client) {
		c.scheme = scheme
	}
}

// Client provides a wrapper around the Namerd HTTP/1.1 API
type Client struct {
	host   string
	scheme string
	client *http.Client
}

// url constructs a url for the path
func (c *Client) url(parts ...string) *url.URL {
	return &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   path.Join(parts...),
	}
}

// Dentries returns the dentries for a specific delegation table
func (c *Client) Dentries(dtab string) (Dentries, error) {
	url := c.url("api", "1", "dtabs", dtab)
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	var dentries Dentries
	dec := json.NewDecoder(rsp.Body)
	if err := dec.Decode(&dentries); err != nil {
		return nil, err
	}
	return dentries, nil
}

// UpdateDentries updates a delegation tables dentries
func (c *Client) UpdateDentries(dtab string, dentries Dentries) error {
	var body = new(bytes.Buffer)
	enc := json.NewEncoder(body)
	if err := enc.Encode(dentries); err != nil {
		return err
	}
	url := c.url("api", "1", "dtabs", dtab)
	req, err := http.NewRequest(http.MethodPut, url.String(), body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	rsp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	switch rsp.StatusCode {
	case http.StatusNoContent:
		return nil
	default:
		return fmt.Errorf("unexpected PUT response code: %s", rsp.StatusCode)
	}
}

// New constructs a new Client, use Option functions to override
// default configuration options
func New(opts ...Option) *Client {
	c := &Client{
		host:   "localhost:4180",
		scheme: "http",
		client: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
