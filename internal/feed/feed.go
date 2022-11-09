package feed

import (
	"dashboard/internal/config"
	"dashboard/internal/wrappers"
	"time"

	"github.com/patrickmn/go-cache"
)

type Feed struct {
	configuredFeeds config.FeedConfig
	feedCache       *cache.Cache

	feedOutgoingMetrics *wrappers.Metrics
}

func New() *Feed {
	return &Feed{
		configuredFeeds:     config.GetFeedConfig(),
		feedCache:           cache.New(5*time.Minute, 10*time.Minute),
		feedOutgoingMetrics: wrappers.NewMetrics("dashboard", "feeds_outgoing"),
	}
}
