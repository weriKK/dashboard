package feed

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mmcdole/gofeed"
	"github.com/patrickmn/go-cache"
)

type feedModel struct {
	Id    int             `json:"id"`
	Name  string          `json:"name"`
	Url   string          `json:"url"`
	Items []feedItemModel `json:"items"`
}

type feedItemModel struct {
	Title     string `json:"title"`
	Url       string `json:"url"`
	Published string `json:"published"`
}

func (f *Feed) GetFeedHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "resource id must be an integer", http.StatusBadRequest)
		return
	}

	if len(f.configuredFeeds) <= id {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	feedItems, err := f.fetchFeedItems(id)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to fetch feed items", http.StatusInternalServerError)
		return
	}

	payload := feedModel{
		Id:    id,
		Name:  f.configuredFeeds[id].Title,
		Url:   f.configuredFeeds[id].WebLink,
		Items: feedItems,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type cachedItem struct {
	items []feedItemModel
	err   error
}

func (f *Feed) fetchFeedItems(id int) ([]feedItemModel, error) {

	cacheKey := strconv.Itoa(id)

	if cachedItems, found := f.feedCache.Get(cacheKey); found {
		ci := cachedItems.(cachedItem)
		if ci.err != nil {
			return nil, fmt.Errorf("cached attempt to fetch feed failed. retrying when cache expires: %w", ci.err)
		}
		return ci.items, nil
	}

	body, err := f.getFeedFromURL(f.configuredFeeds[id].FeedLink)
	if err != nil {
		f.feedCache.Set(cacheKey, cachedItem{items: nil, err: err}, cache.DefaultExpiration)
		return nil, err
	}

	p := gofeed.NewParser()

	parsed, err := p.ParseString(string(body))
	if err != nil {
		emsg := fmt.Errorf("failed to parse feed: %w", err)
		f.feedCache.Set(cacheKey, cachedItem{items: nil, err: emsg}, cache.DefaultExpiration)
		return nil, emsg
	}

	limit := int(math.Min(float64(f.configuredFeeds[id].ItemLimit), float64(len(parsed.Items))))

	var items []feedItemModel
	for i := 0; i < limit; i++ {
		items = append(items, feedItemModel{
			Title:     parsed.Items[i].Title,
			Url:       parsed.Items[i].Link,
			Published: parsed.Items[i].Published,
		})
	}

	f.feedCache.Set(cacheKey, cachedItem{items: items, err: nil}, cache.DefaultExpiration)
	return items, nil
}
