// Package rssclient provides http client for requesting rss feeds
package rssclient

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration) Client {
	return Client{
		&http.Client{
			Timeout: timeout,
		},
	}
}

type statusError struct {
	code int
	body string
}

func (e *statusError) Error() string {
	return fmt.Sprintf("response failed with status code %d: %s", e.code, e.body)
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
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("could not read response body: %w", err)
	}

	if res.StatusCode > 299 {
		return []byte{}, &statusError{res.StatusCode, string(body)}
	}

	return body, nil
}

type Channel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

type RSSFeed struct {
	Channel Channel `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (r *RSSFeed) RemoveHTMLUnescape() {
	r.Channel.Title = html.UnescapeString(r.Channel.Title)
	r.Channel.Link = html.UnescapeString(r.Channel.Link)
	r.Channel.Description = html.UnescapeString(r.Channel.Description)

	for i := range r.Channel.Items {
		item := &r.Channel.Items[i]

		item.Title = html.UnescapeString(item.Title)
		item.Link = html.UnescapeString(item.Link)
		item.Description = html.UnescapeString(item.Description)
		item.PubDate = html.UnescapeString(item.PubDate)
	}
}

func (c Client) FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	data, err := c.requestFromURL(ctx, feedURL)
	if err != nil {
		return nil, err
	}

	var RSSFeed RSSFeed
	err = xml.Unmarshal(data, &RSSFeed)
	if err != nil {
		return nil, err
	}

	if RSSFeed.Channel.Items == nil {
		RSSFeed.Channel.Items = []RSSItem{}
	}

	return &RSSFeed, nil
}
