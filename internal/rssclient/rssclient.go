// Package rssclient provides http client for requesting rss feeds
package rssclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient http.Client
}

func NewClient(timeout time.Duration) Client {
	return Client{
		http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) requestFromURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("could not creat request: %w", err)
	}
	req.Header.Set("User-Agent", "gator")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("could not fetch data: %w", err)
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			fmt.Println("could not close response body: ", err)
		}
	}()

	body, err := io.ReadAll(res.Body)
	if res.StatusCode > 299 {
		return []byte{}, fmt.Errorf("response failed with status code %d: %s", res.StatusCode, string(body))
	}
	if err != nil {
		return []byte{}, fmt.Errorf("could not read response body: %w", err)
	}

	return body, nil
}
