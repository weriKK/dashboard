package backend

import (
	"sort"
	"strings"
	"time"
)

func ScoreItem(item *FeedItem) float64 {
	return tokenAffinityScore(Tokenize(item.Title))
}

// tokenAffinityScore computes the dot product between an item's tokens and
// the user's learned token preference weights. Higher means more relevant
// to what the user has clicked before.
func tokenAffinityScore(words []string) float64 {
	TokenWeightMu.RLock()
	defer TokenWeightMu.RUnlock()

	score := 0.0
	for _, w := range words {
		if weight, ok := TokenWeights[w]; ok {
			score += weight
		}
	}
	return score
}

// GetTopRatedItems returns strict top-N scored items globally across all feeds.
func GetTopRatedItems(limit int) []TopRatedItem {
	if limit <= 0 {
		return []TopRatedItem{}
	}

	TokenWeightMu.RLock()
	hasWeights := len(TokenWeights) > 0
	TokenWeightMu.RUnlock()
	if !hasWeights {
		return []TopRatedItem{}
	}

	allFeeds := GetAllFeeds()

	var allItems []FeedItem
	maxAgeHours := time.Duration(Cfg.ML.MaxItemAgeHours) * time.Hour
	now := time.Now()

	for _, group := range allFeeds {
		for _, item := range group.Items {
			if now.Sub(item.PublishedAt) <= maxAgeHours {
				allItems = append(allItems, item)
			}
		}
	}

	type scoredItem struct {
		item  FeedItem
		score float64
	}

	scored := make([]scoredItem, 0, len(allItems))
	for _, item := range allItems {
		scored = append(scored, scoredItem{item: item, score: ScoreItem(&item)})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	if len(scored) > limit {
		scored = scored[:limit]
	}

	result := make([]TopRatedItem, 0, len(scored))
	for _, s := range scored {
		result = append(result, TopRatedItem{
			Link:  s.item.Link,
			Score: s.score,
		})
	}

	return result
}

// Tokenize splits text into meaningful words
func Tokenize(text string) []string {
	text = strings.ToLower(text)
	// Simple tokenization: split on whitespace and punctuation
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})

	// Filter out very common stopwords
	stopwords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"to": true, "in": true, "on": true, "at": true, "by": true,
		"for": true, "of": true, "with": true, "is": true, "are": true,
		"be": true, "it": true, "as": true, "was": true, "were": true,
	}

	filtered := []string{}
	for _, w := range words {
		if len(w) > 2 && !stopwords[w] {
			filtered = append(filtered, w)
		}
	}
	return filtered
}
