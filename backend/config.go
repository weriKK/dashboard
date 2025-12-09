package backend

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadConfig reads and parses the YAML configuration file
func LoadConfig() error {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return fmt.Errorf("error reading config.yaml: %w", err)
	}

	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		return fmt.Errorf("error parsing config.yaml: %w", err)
	}

	// Validate required fields
	if Cfg.Stocks.API.APIKey == "" || Cfg.Stocks.API.APIKey == "YOUR_FINNHUB_API_KEY_HERE" {
		log.Println("WARNING: Finnhub API key not configured. Stock data will not be available.")
	}

	return nil
}

// InitFeedCache initializes the feed cache with entries for all configured sources
func InitFeedCache() {
	FeedCache = make(map[string]*FeedCacheEntry)
	for _, category := range Cfg.Feeds {
		for _, source := range category.Sources {
			cacheKey := fmt.Sprintf("%s:%s", category.Category, source.Name)
			FeedCache[cacheKey] = &FeedCacheEntry{
				Items:       []*FeedItem{},
				NextRefresh: time.Now(),
				Interval:    time.Duration(Cfg.Refresh.IntervalMinutes) * time.Minute,
			}
		}
	}
	StockCache = make(map[string]*StockData)
	ClickHistory = []ClickFeedback{}
}
