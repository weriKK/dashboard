package infrastructure

import (
	"errors"
	"sort"
	"sync"

	"github.com/weriKK/dashboard/domain/feed"
)

type entry struct {
	name   string
	url    string
	rss    string
	column int
	limit  int
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
		newFeed := feed.New(k, v.name, v.url, v.rss, v.column, v.limit)
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
			newFeed = feed.New(id, db.data[id].name, db.data[id].url, db.data[id].rss, db.data[id].column, db.data[id].limit)
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
	// Column 0
	db.add(&entry{"MMO-Champion", "https://www.mmo-champion.com", "http://www.mmo-champion.com/external.php?do=rss&type=newcontent&sectionid=1&days=120&count=20", 0, 5})
	db.add(&entry{"Reddit - Games", "https://www.reddit.com/r/Games/", "https://www.reddit.com/r/Games/.rss", 0, 12})
	db.add(&entry{"RockPaperShotgun", "https://www.rockpapershotgun.com", "http://feeds.feedburner.com/RockPaperShotgun", 0, 5})
	db.add(&entry{"Bluenews", "https://www.bluesnews.com", "http://www.bluesnews.com/news/news_1_0.rdf", 0, 8})
	
	// Column 1
	db.add(&entry{"Jason Schreier", "https://www.bloomberg.com/authors/AUvqMRVAZCw/jason-schreier", "https://www.bloomberg.com/authors/AUvqMRVAZCw/jason-schreier.rss", 1, 10})
	
	// Column 2
	db.add(&entry{"ArsTechnica", "https://arstechnica.com", "http://feeds.arstechnica.com/arstechnica/index", 2, 10})
	db.add(&entry{"Y Combinator", "https://news.ycombinator.com", "https://news.ycombinator.com/rss", 2, 20})
}

func NewInMemoryFeedRepository() *inMemoryFeedRepository {
	db := inMemoryFeedRepository{data: make(map[int]*entry)}
	db.initializeWithData()
	return &db
}
