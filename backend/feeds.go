package backend

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/mmcdole/gofeed"
)

// FetchFeed retrieves and parses an RSS feed with conditional request headers and retries
func FetchFeed(ctx context.Context, source FeedSource, category string) (*FeedCacheEntry, error) {
	cacheKey := fmt.Sprintf("%s:%s", category, source.Name)
	entry := FeedCache[cacheKey]

	backoffs := []time.Duration{1 * time.Second, 3 * time.Second, 9 * time.Second}
	client := &http.Client{Timeout: 20 * time.Second}

	var lastErr error
	for attempt := 0; attempt < len(backoffs); attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
		if err != nil {
			return entry, err
		}

		// Add conditional request headers for bandwidth efficiency
		if entry.ETag != "" {
			req.Header.Set("If-None-Match", entry.ETag)
		}
		if entry.LastModified != "" {
			req.Header.Set("If-Modified-Since", entry.LastModified)
		}

		// Browser-like user agent to reduce blocks
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < len(backoffs)-1 {
				time.Sleep(backoffs[attempt])
				continue
			}
			return entry, lastErr
		}

		// Ensure body closed per attempt
		func() {
			defer resp.Body.Close()

			// Handle 304 Not Modified
			if resp.StatusCode == http.StatusNotModified {
				newInterval := time.Duration(float64(entry.Interval) * Cfg.Refresh.NotModifiedBackoffMultiplier)
				maxInterval := time.Duration(Cfg.Refresh.MaxIntervalMinutes) * time.Minute
				if newInterval > maxInterval {
					newInterval = maxInterval
				}
				entry.Interval = newInterval
				entry.NextRefresh = time.Now().Add(newInterval)
				lastErr = nil
				return
			}

			// Non-200 is treated as failure
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				lastErr = fmt.Errorf("unexpected status %d from %s", resp.StatusCode, source.URL)
				return
			}

			// Reset to default interval on successful fetch
			entry.Interval = time.Duration(Cfg.Refresh.IntervalMinutes) * time.Minute

			// Update cache headers
			if etag := resp.Header.Get("ETag"); etag != "" {
				entry.ETag = etag
			}
			if lastMod := resp.Header.Get("Last-Modified"); lastMod != "" {
				entry.LastModified = lastMod
			}

			// Parse feed with gofeed (handles RSS, Atom, and other formats)
			fp := gofeed.NewParser()
			parsedFeed, err := fp.Parse(resp.Body)
			if err != nil {
				lastErr = fmt.Errorf("error parsing feed: %w", err)
				return
			}

			items := []*FeedItem{}
			for _, item := range parsedFeed.Items {
				if item == nil {
					continue
				}

				pubTime := time.Now()
				if item.PublishedParsed != nil {
					pubTime = *item.PublishedParsed
				}

				items = append(items, &FeedItem{
					Title:       item.Title,
					Link:        item.Link,
					PublishedAt: pubTime,
					Source:      source.Name,
					Category:    category,
					Description: item.Description,
					GUID:        item.GUID,
				})
			}

			if len(items) == 0 {
				log.Printf("Warning: No items parsed from %s (%s)", source.Name, source.URL)
			} else {
				log.Printf("Parsed %d items from %s", len(items), source.Name)
			}

			// Sort by published date, newest first
			sort.Slice(items, func(i, j int) bool {
				return items[i].PublishedAt.After(items[j].PublishedAt)
			})

			entry.Items = items
			entry.LastFetch = time.Now()
			entry.NextRefresh = time.Now().Add(entry.Interval)
			lastErr = nil
		}()

		if lastErr == nil {
			return entry, nil
		}

		if attempt < len(backoffs)-1 {
			time.Sleep(backoffs[attempt])
		}
	}

	// If we exhausted retries, push next refresh back slightly and return last error
	if entry.NextRefresh.Before(time.Now().Add(entry.Interval)) {
		entry.NextRefresh = time.Now().Add(entry.Interval)
	}

	return entry, lastErr
}

// RefreshFeedsWorker continuously refreshes feeds according to their schedule
func RefreshFeedsWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			for _, category := range Cfg.Feeds {
				for _, source := range category.Sources {
					cacheKey := fmt.Sprintf("%s:%s", category.Category, source.Name)
					entry := FeedCache[cacheKey]

					if now.After(entry.NextRefresh) {
						log.Printf("Refreshing feed: %s:%s", category.Category, source.Name)
						FeedCacheMu.Lock()
						_, err := FetchFeed(ctx, source, category.Category)
						if err != nil {
							log.Printf("Error fetching %s: %v", source.Name, err)
						}
						FeedCache[cacheKey] = entry
						FeedCacheMu.Unlock()
					}
				}
			}
		}
	}
}

// GetAllFeeds retrieves all available feeds grouped by source
func GetAllFeeds() []FeedGroup {
	FeedCacheMu.RLock()
	defer FeedCacheMu.RUnlock()

	var result []FeedGroup

	for _, category := range Cfg.Feeds {
		for _, source := range category.Sources {
			cacheKey := fmt.Sprintf("%s:%s", category.Category, source.Name)
			group := FeedGroup{Source: source.Name, Category: category.Category, Color: category.Color, SiteURL: source.Site, Items: []FeedItem{}}

			if entry, ok := FeedCache[cacheKey]; ok {
				for _, item := range entry.Items {
					item.Score = 0 // Will be scored later if needed
					group.Items = append(group.Items, *item)
				}
			}

			// Sort by date, newest first
			sort.Slice(group.Items, func(i, j int) bool {
				return group.Items[i].PublishedAt.After(group.Items[j].PublishedAt)
			})

			result = append(result, group)
		}
	}

	return result
}
