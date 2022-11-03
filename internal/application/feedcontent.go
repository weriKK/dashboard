package application

import (
	"github.com/mmcdole/gofeed"
	log "dashboard/internal/util/logger"
)

type FeedItem struct {
	Title       string
	Url         string
	Published   string
}

type FeedContent struct {
	Id     int
	Name   string
	Url    string
	Column int
	Items  []FeedItem
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetFeedContent(id int, limit int) (*FeedContent, error) {

	feed, _ := appdb.GetById(id)

	feedParser := gofeed.NewParser()
	parsed, err := feedParser.ParseURL(feed.Rss())
	if err != nil {
		log.Panicf("Exception while trying to parse RSS feed URL '%v': %v", feed.Rss(), err)
	}

	if limit == 0 {
		limit = len(parsed.Items)
	}

	limit = min(limit, len(parsed.Items))

	items := []FeedItem{}
	for itemIdx := 0; itemIdx < limit; itemIdx++ {
		p := parsed.Items[itemIdx]
		items = append(items, FeedItem{p.Title, p.Link, p.Published})
	}

	content := FeedContent{feed.Id(), parsed.Title, feed.Url(), feed.Column(), items}
	return &content, nil
}
