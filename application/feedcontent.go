package application

import "github.com/mmcdole/gofeed"

type FeedItem struct {
	Title       string
	Url         string
	Description string
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

func GetFeedContent(id int) (*FeedContent, error) {

	feed, _ := appdb.GetById(id)

	feedParser := gofeed.NewParser()
	parsed, err := feedParser.ParseURL(feed.Rss())
	if err != nil {
		panic(err)
	}

	items := []FeedItem{}
	for _, v := range parsed.Items {
		items = append(items, FeedItem{v.Title, v.Link, v.Description})
	}

	// TODO: this is a hack, fix the API so returned items can be limited/configured
	limitItemsLength := min(len(items), 8)

	content := FeedContent{feed.Id(), parsed.Title, feed.Url(), feed.Column(), items[:limitItemsLength]}
	return &content, nil
}
