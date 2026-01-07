package rssclient

import (
	"context"
	"testing"
	"time"
)

var (
	client  = NewClient(5 * time.Second)
	testURL = "https://www.wagslane.dev/index.xml"
)

func TestHttpGet(t *testing.T) {
	_, err := client.requestFromURL(context.Background(), testURL)
	if err != nil {
		t.Log("Error: ", err)
	}
}

func TestFetchFeed(t *testing.T) {
	actual, err := client.FetchFeed(context.Background(), testURL)
	if err != nil {
		t.Log("Error: ", err)
	}

	t.Log(actual)
}
