package backend

import (
	"sync"
	"time"
)

// Config structures
type Config struct {
	Server    ServerConfig     `yaml:"server"`
	Stocks    StocksConfig     `yaml:"stocks"`
	Feeds     []FeedCategory   `yaml:"feeds"`
	Timezones []TimezoneConfig `yaml:"timezones"`
	Refresh   RefreshConfig    `yaml:"refresh"`
	ML        MLConfig         `yaml:"ml"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type FinnhubConfig struct {
	APIKey  string `yaml:"apiKey" json:"apiKey"`
	BaseURL string `yaml:"baseURL" json:"baseURL"`
}

type StockConfig struct {
	Symbol       string `yaml:"symbol"`
	Exchange     string `yaml:"exchange"`
	BaseCurrency string `yaml:"baseCurrency"`
	Label        string `yaml:"label"`
}

type StocksConfig struct {
	API     FinnhubConfig `yaml:"api"`
	Items   []StockConfig `yaml:"items"`
	Enabled bool          `yaml:"enabled"`
	Refresh int           `yaml:"refreshIntervalHours"`
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

type TimezoneConfig struct {
	Name   string `yaml:"name" json:"name"`
	City   string `yaml:"city" json:"city"`
	Offset string `yaml:"offset" json:"offset"`
}

type RefreshConfig struct {
	IntervalMinutes              int     `yaml:"intervalMinutes"`
	NotModifiedBackoffMultiplier float64 `yaml:"notModifiedBackoffMultiplier"`
	MaxIntervalMinutes           int     `yaml:"maxIntervalMinutes"`
}

type MLConfig struct {
	MaxItemAgeHours          int         `yaml:"maxItemAgeHours"`
	DiversitySamplingPercent int         `yaml:"diversitySamplingPercent"`
	RecommendationCount      int         `yaml:"recommendationCount"`
	TFIDF                    TFIDFConfig `yaml:"tfidf"`
	ClickDecayPerDay         float64     `yaml:"clickDecayPerDay"`
	ClickWeight              float64     `yaml:"clickWeight"`
}

type TFIDFConfig struct {
	MinDocFreq int `yaml:"minDocFreq"`
	MaxDocFreq int `yaml:"maxDocFreq"`
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

type StockData struct {
	Symbol        string    `json:"symbol"`
	Label         string    `json:"label"`
	Price         float64   `json:"price"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"changePercent"`
	Trend         []int     `json:"trend"` // Last 7 days high/low trend
	UpdatedAt     time.Time `json:"updatedAt"`
}

// FeedGroup represents a single feed source and its items
type FeedGroup struct {
	Source   string     `json:"source"`
	Category string     `json:"category"`
	Color    string     `json:"color"`
	SiteURL  string     `json:"siteUrl"`
	Items    []FeedItem `json:"items"`
}

type RecommendedItem struct {
	Title  string  `json:"title"`
	Link   string  `json:"link"`
	Age    string  `json:"age"`
	Source string  `json:"source"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type APIResponse struct {
	Feeds           []FeedGroup       `json:"feeds"`
	Stocks          []StockData       `json:"stocks"`
	Recommendations []RecommendedItem `json:"recommendations"`
	Timezones       []TimezoneConfig  `json:"timezones"`
	CurrentTime     time.Time         `json:"currentTime"`
}

type ClickFeedback struct {
	ItemGUID  string    `json:"itemGUID"`
	ItemTitle string    `json:"itemTitle"`
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
	FeedCache      map[string]*FeedCacheEntry
	FeedCacheMu    sync.RWMutex
	StockCache     map[string]*StockData
	StockCacheMu   sync.RWMutex
	ClickHistory   []ClickFeedback
	ClickHistoryMu sync.RWMutex
	Cfg            Config
)
