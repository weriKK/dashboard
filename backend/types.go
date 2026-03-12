package backend

import (
	"sync"
	"time"
)

// Config structures
type Config struct {
	Server  ServerConfig   `yaml:"server"`
	Feeds   []FeedCategory `yaml:"feeds"`
	Refresh RefreshConfig  `yaml:"refresh"`
	ML      MLConfig       `yaml:"ml"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type FeedCategory struct {
	Category string       `yaml:"category"`
	Color    string       `yaml:"color"`
	Sources  []FeedSource `yaml:"sources"`
}

type FeedSource struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Site string `yaml:"siteUrl" json:"siteUrl"`
}

type RefreshConfig struct {
	IntervalMinutes              int     `yaml:"intervalMinutes"`
	NotModifiedBackoffMultiplier float64 `yaml:"notModifiedBackoffMultiplier"`
	MaxIntervalMinutes           int     `yaml:"maxIntervalMinutes"`
}

type MLConfig struct {
	MaxItemAgeHours  int     `yaml:"maxItemAgeHours"`
	ClickWeight      float64 `yaml:"clickWeight"`
	TokenDecayPerDay float64 `yaml:"tokenDecayPerDay"`
	DBPath           string  `yaml:"dbPath"`
	RetentionDays    int     `yaml:"retentionDays"`
}

// Domain models
type FeedItem struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	PublishedAt time.Time `json:"publishedAt"`
	Source      string    `json:"source"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	GUID        string    `json:"guid"`
	Score       float64   `json:"score"`
	Age         string    `json:"age"`
}

// FeedGroup represents a single feed source and its items
type FeedGroup struct {
	Source   string     `json:"source"`
	Category string     `json:"category"`
	Color    string     `json:"color"`
	SiteURL  string     `json:"siteUrl"`
	Items    []FeedItem `json:"items"`
}

type TopRatedItem struct {
	Link  string  `json:"link"`
	Score float64 `json:"score"`
}

type APIResponse struct {
	Feeds    []FeedGroup    `json:"feeds"`
	TopRated []TopRatedItem `json:"topRated"`
}

type ClickFeedback struct {
	ItemKey   string    `json:"itemKey"`
	ItemTitle string    `json:"itemTitle"`
	ItemLink  string    `json:"itemLink"`
	Source    string    `json:"source"`
	Category  string    `json:"category"`
	Timestamp time.Time `json:"timestamp"`
}

// FeedCacheEntry tracks cached feed data
type FeedCacheEntry struct {
	Items        []*FeedItem
	ETag         string
	LastModified string
	LastFetch    time.Time
	NextRefresh  time.Time
	Interval     time.Duration
}

// Global state
var (
	FeedCache   map[string]*FeedCacheEntry
	FeedCacheMu sync.RWMutex
	Cfg         Config
)
