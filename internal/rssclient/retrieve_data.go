package rssclient

import (
	"context"
	"encoding/xml"
	"html"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Items       []RSSItem `xml:"item"`
	} `xml:"channel"`
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

	for _, item := range r.Channel.Items {
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

	return &RSSFeed, nil
}
