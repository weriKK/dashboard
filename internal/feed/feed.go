package feed

import (
	"dashboard/internal/config"
	"time"

	"github.com/patrickmn/go-cache"
)

type Feed struct {
	configuredFeeds config.FeedConfig
	feedCache       *cache.Cache
}

func New() *Feed {
	return &Feed{
		configuredFeeds: config.GetFeedConfig(),
		feedCache:       cache.New(5*time.Minute, 10*time.Minute),
	}
}
