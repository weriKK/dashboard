package feed

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type feedListModel struct {
	Feeds []feedListItemModel
}

type feedListItemModel struct {
	Name     string
	Url      string
	Resource string
	Column   int
}

func (f *Feed) GetFeedListHandler(w http.ResponseWriter, r *http.Request) {

	payload := feedListModel{}

	for i, feed := range f.configuredFeeds {
		payload.Feeds = append(payload.Feeds, feedListItemModel{
			Name:     feed.Title,
			Url:      feed.WebLink,
			Resource: fmt.Sprintf("https://%s/webfeeds/%d", r.Host, i),
			Column:   feed.ColumnId,
		})
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
