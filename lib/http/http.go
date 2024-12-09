package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
)

// Client is a wrapper around the Go stdlib http client.
type Client struct {
	internal *http.Client
}

// NewClient returns a new Client with its internal http client
// set to the default client.
func NewClient() *Client {
	return &Client{
		internal: http.DefaultClient,
	}
}

type GetOptions func(*http.Request)

func WithHeader(key, value string) GetOptions {
	return func(r *http.Request) {
		r.Header.Add(key, value)
	}
}

func WithJSONAccept() GetOptions {
	return func(r *http.Request) {
		r.Header.Add("Accept", "application/json")
	}
}

func WithQueryParam(key, value string) GetOptions {
	return func(r *http.Request) {
		q := r.URL.Query()
		q.Add(key, value)
		r.URL.RawQuery = q.Encode()
	}
}

// GetWithContext performs a Get request with the context provided.
func (c *Client) GetWithContext(ctx context.Context, url string, opts ...GetOptions) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, err := retry.DoWithData(func() (*http.Response, error) {
		resp, err := c.internal.Do(req)
		if err != nil {
			return nil, retry.Unrecoverable(err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, checkResponseCode(resp)
		}

		return resp, nil
	}, retry.Attempts(20), retry.Delay(1*time.Second))
	if err != nil {
		return resp, err
	}

	return resp, checkResponseCode(resp)
}

// checkResponseCode parses the http.Response status code and returns
// populated errors based on the code.
func checkResponseCode(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("%s %s %d not found", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)
	case http.StatusUnauthorized:
		return fmt.Errorf("%s %s %d unauthorized", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%s %s %d too many requests", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)
	case http.StatusNotImplemented:
		return fmt.Errorf("%s %s %d not implemented", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)
	case http.StatusBadGateway:
		return fmt.Errorf("%s %s %d bad gateway", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)
	case http.StatusServiceUnavailable:
		return fmt.Errorf("%s %s %d service unavailable", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)
	case http.StatusGatewayTimeout:
		return fmt.Errorf("%s %s %d gateway timeout", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)
	case http.StatusHTTPVersionNotSupported:
		return fmt.Errorf("%s %s %d unsupported HTTP version", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)
	case http.StatusTeapot:
		return fmt.Errorf("%s %s %d teapot", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)
	case http.StatusInternalServerError:
		return fmt.Errorf("%s %s %d internal server error", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)
	case http.StatusForbidden:
		return fmt.Errorf("%s %s %d forbidden", resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
