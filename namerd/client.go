package namerd

import (
	"fmt"
	"net/http"
)

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
