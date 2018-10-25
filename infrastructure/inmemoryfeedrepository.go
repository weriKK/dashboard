package infrastructure

import (
	"errors"
	"github.com/weriKK/dashboard/domain/feed"
	"sort"
	"sync"
)

type entry struct {
	name  string
	url   string
	limit int
}

type inMemoryFeedRepository struct {
	data map[int]*entry
	mux  sync.RWMutex
}

func (db *inMemoryFeedRepository) GetAll(limit int) ([]feed.Feed, error) {

	var data []feed.Feed
	count := 0

	for k, v := range db.data {
		if limit <= count {
			break
		}
		newFeed := feed.New(k, v.name, v.url, v.limit)
		data = append(data, newFeed)
		count++
	}

	// Fun fact, map stores elements in a randomized order,
	// can't rely on it keeping indices in the expected order
	sort.Slice(data, func(i, j int) bool {
		return data[i].Id() < data[j].Id()
	})

	return data, nil
}

func (db *inMemoryFeedRepository) GetById(id int) (*feed.Feed, error) {

	newFeed := feed.Feed{}

	for k := range db.data {
		if k == id {
			newFeed = feed.New(id, db.data[id].name, db.data[id].url, db.data[id].limit)
			return &newFeed, nil
		}
	}

	return &newFeed, errors.New("Feed with given id not found!")

}

func (db *inMemoryFeedRepository) Count() (int, error) {
	return len(db.data), nil
}

func (db *inMemoryFeedRepository) add(value *entry) {
	db.mux.Lock()
	db.data[len(db.data)] = value
	db.mux.Unlock()
}

func (db *inMemoryFeedRepository) initializeWithData() {
	db.add(&entry{"MMO-Champion", "http://www.mmo-champion.com/external.php?do=rss&type=newcontent&sectionid=1&days=120&count=20", 10})
	db.add(&entry{"Reddit - Games", "https://www.reddit.com/r/Games/.rss", 10})
	db.add(&entry{"Programming Praxis", "https://programmingpraxis.com/feed/", 10})
	db.add(&entry{"Handmade Hero", "https://www.youtube.com/feeds/videos.xml?channel_id=UCaTznQhurW5AaiYPbhEA-KA", 10})
	db.add(&entry{"GiantBomb", "http://www.giantbomb.com/feeds/mashup/", 10})
	db.add(&entry{"RockPaperShotgun", "http://feeds.feedburner.com/RockPaperShotgun", 10})
	db.add(&entry{"Shacknews", "http://www.shacknews.com/rss?recent_articles=1", 10})
	db.add(&entry{"Bluenews", "http://www.bluesnews.com/news/news_1_0.rdf", 10})
	db.add(&entry{"Gamasutra", "http://feeds.feedburner.com/GamasutraFeatureArticles/", 10})
	db.add(&entry{"ArsTechnica", "http://feeds.arstechnica.com/arstechnica/index", 10})
	db.add(&entry{"GamesIndustry", "http://www.gamesindustry.biz/rss/gamesindustry_news_feed.rss", 10})
	db.add(&entry{"Y Combinator", "https://news.ycombinator.com/rss", 10})
}

func NewInMemoryFeedRepository() *inMemoryFeedRepository {
	db := inMemoryFeedRepository{data: make(map[int]*entry)}
	db.initializeWithData()
	return &db
}
