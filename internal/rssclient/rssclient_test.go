package rssclient

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestRemoveHTMLUnescape(t *testing.T) {
	item1 := RSSItem{
		Title:       "&lt;title&gt; test item 1",
		Link:        "&lt;link&gt; test.com/item1",
		Description: "&lt;description&gt; item 1 test description",
		PubDate:     "&lt;timedate&gt; 01/01/01",
	}
	item2 := RSSItem{
		Title:       "test item 2",
		Link:        "test.com/item2",
		Description: "item 2 test description",
		PubDate:     "02/02/02",
	}

	feed := &RSSFeed{
		Channel: Channel{
			Title:       "&lt;title&gt; Test Title",
			Link:        "&lt;title&gt; test.com",
			Description: "&lt;title&gt; test description",
			Items:       []RSSItem{item1, item2},
		},
	}

	feed.RemoveHTMLUnescape()

	expectedItem1 := RSSItem{
		Title:       "<title> test item 1",
		Link:        "<link> test.com/item1",
		Description: "<description> item 1 test description",
		PubDate:     "<timedate> 01/01/01",
	}
	expectedFeed := &RSSFeed{
		Channel: Channel{
			Title:       "<title> Test Title",
			Link:        "<title> test.com",
			Description: "<title> test description",
			Items:       []RSSItem{expectedItem1, item2},
		},
	}

	if !reflect.DeepEqual(feed, expectedFeed) {
		t.Errorf("expected %v but got %v", expectedFeed, feed)
	}
}

func TestRequestFromURL(t *testing.T) {
	type testCase struct {
		name        string
		status      int
		expectedErr reflect.Type
		body        string
	}

	testCases := []testCase{
		{
			name:        "valid response test",
			status:      200,
			expectedErr: nil,
			body:        "<rss>here is an rss feed</rss>",
		},
		{
			name:        "not found test",
			status:      404,
			expectedErr: reflect.TypeOf(&statusError{}),
			body:        "test",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(testCase.status)

					_, err := w.Write([]byte(testCase.body))
					if err != nil {
						t.Fatalf("Server unable to write %v to response", testCase.body)
					}

					if userAgent := r.Header.Get("User-Agent"); userAgent != "gator" {
						t.Fatal("expected 'User-Agent' to be set to 'gator' in header")
					}
				}))
			defer server.Close()

			client := NewClient(5 * time.Second)
			response, err := client.requestFromURL(context.Background(), server.URL)
			if testCase.expectedErr != nil {
				if reflect.TypeOf(err) == testCase.expectedErr {
					return
				} else {
					t.Fatalf("expected error %v but got error %v", testCase.expectedErr, err)
				}
			}

			responseString := string(response)
			if testCase.expectedErr == nil {
				if testCase.body != responseString {
					t.Fatalf("expected body %v but got %v", testCase.body, responseString)
				}
			}
		})
	}
}

func TestFetchFeed(t *testing.T) {
	type testCase struct {
		name         string
		rawRSS       string
		expectedFeed *RSSFeed
	}

	testCases := []testCase{
		{
			name: "basic unmarshal",
			rawRSS: `
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Example Feed</title>
    <link>https://example.com</link>
    <description>An example RSS feed</description>
    <item>
      <title>First item</title>
      <link>https://example.com/item1</link>
      <description>Item 1 description</description>
      <pubDate>Mon, 01 Jan 2024 00:00:00 GMT</pubDate>
    </item>
  </channel>
</rss>
`,
			expectedFeed: &RSSFeed{
				Channel: Channel{
					Title:       "Example Feed",
					Link:        "https://example.com",
					Description: "An example RSS feed",
					Items: []RSSItem{
						{
							Title:       "First item",
							Link:        "https://example.com/item1",
							Description: "Item 1 description",
							PubDate:     "Mon, 01 Jan 2024 00:00:00 GMT",
						},
					},
				},
			},
		},
		{
			name: "multiple items",
			rawRSS: `
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Multi Item Feed</title>
    <link>https://example.com</link>
    <description>Feed with multiple items</description>
    <item>
      <title>Item One</title>
      <link>https://example.com/1</link>
      <description>First item</description>
      <pubDate>Mon, 01 Jan 2024 00:00:00 GMT</pubDate>
    </item>
    <item>
      <title>Item Two</title>
      <link>https://example.com/2</link>
      <description>Second item</description>
      <pubDate>Tue, 02 Jan 2024 00:00:00 GMT</pubDate>
    </item>
  </channel>
</rss>
`,
			expectedFeed: &RSSFeed{
				Channel: Channel{
					Title:       "Multi Item Feed",
					Link:        "https://example.com",
					Description: "Feed with multiple items",
					Items: []RSSItem{
						{
							Title:       "Item One",
							Link:        "https://example.com/1",
							Description: "First item",
							PubDate:     "Mon, 01 Jan 2024 00:00:00 GMT",
						},
						{
							Title:       "Item Two",
							Link:        "https://example.com/2",
							Description: "Second item",
							PubDate:     "Tue, 02 Jan 2024 00:00:00 GMT",
						},
					},
				},
			},
		},
		{
			name: "no items",
			rawRSS: `
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Empty Feed</title>
    <link>https://example.com</link>
    <description>No items yet</description>
  </channel>
</rss>
`,
			expectedFeed: &RSSFeed{
				Channel: Channel{
					Title:       "Empty Feed",
					Link:        "https://example.com",
					Description: "No items yet",
					Items:       []RSSItem{},
				},
			},
		},
		{
			name: "partial rss",
			rawRSS: `
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Partial Feed</title>
    <link>https://example.com</link>
    <item>
      <title>Item without extras</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>
`,
			expectedFeed: &RSSFeed{
				Channel: Channel{
					Title: "Partial Feed",
					Link:  "https://example.com",
					Items: []RSSItem{
						{
							Title: "Item without extras",
							Link:  "https://example.com/item",
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					_, err := w.Write([]byte(testCase.rawRSS))
					if err != nil {
						t.Fatalf("Server unable to write %v to response", testCase.rawRSS)
					}

					if userAgent := r.Header.Get("User-Agent"); userAgent != "gator" {
						t.Fatal("expected 'User-Agent' to be set to 'gator' in header")
					}
				}))
			defer server.Close()

			client := NewClient(5 * time.Second)
			actual, err := client.FetchFeed(context.Background(), server.URL)
			if err != nil {
				t.Fatalf("could not fetch feed: %v", err)
			}

			if !reflect.DeepEqual(actual, testCase.expectedFeed) {
				log.Fatalf("expected %v but got %v", testCase.expectedFeed, actual)
			}
		})
	}
}
