package application

import "github.com/mmcdole/gofeed"
import "log"

type FeedItem struct {
	Title   string
	Url     string
	Content string
}

type FeedContent struct {
	Id    int
	Name  string
	Url   string
	Items []FeedItem
}

func GetFeedContent(id int) (*FeedContent, error) {

	feed, _ := appdb.GetById(id)

	feedParser := gofeed.NewParser()
	parsed, err := feedParser.ParseURL(feed.Url())
	if err != nil {
		panic(err)
	}

	items := []FeedItem{}
	for _, v := range parsed.Items {
		items = append(items, FeedItem{v.Title, v.Link, v.Description})
	}

	content := FeedContent{feed.Id(), parsed.Title, feed.Url(), items}
	return &content, nil
}
