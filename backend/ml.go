package backend

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// ScoreItem calculates a TF-IDF + click-decay score for a feed item
func ScoreItem(item *FeedItem, allItems []FeedItem) float64 {
	words := Tokenize(item.Title)

	// Count word frequencies across all items
	docFreq := make(map[string]int)
	for _, otherItem := range allItems {
		otherWords := Tokenize(otherItem.Title)
		uniqueWords := make(map[string]bool)
		for _, w := range otherWords {
			uniqueWords[w] = true
		}
		for w := range uniqueWords {
			docFreq[w]++
		}
	}

	// Calculate TF-IDF score
	score := 0.0
	wordFreq := make(map[string]int)
	for _, w := range words {
		wordFreq[w]++
	}

	totalDocs := float64(len(allItems))
	for word, freq := range wordFreq {
		df := float64(docFreq[word])
		if df < float64(Cfg.ML.TFIDF.MinDocFreq) || df > float64(Cfg.ML.TFIDF.MaxDocFreq) {
			continue
		}

		tf := float64(freq) / float64(len(words))
		idf := math.Log(totalDocs / (df + 1))
		score += tf * idf
	}

	// Add click feedback score
	ClickHistoryMu.RLock()
	for _, click := range ClickHistory {
		if click.ItemTitle == item.Title {
			// Decay based on age
			daysSince := time.Since(click.Timestamp).Hours() / 24
			decayFactor := math.Pow(Cfg.ML.ClickDecayPerDay, daysSince)
			score += Cfg.ML.ClickWeight * decayFactor
		}
	}
	ClickHistoryMu.RUnlock()

	return score
}

// GetRecommendations returns ML-ranked recommendations with diversity sampling
func GetRecommendations() []RecommendedItem {
	allFeeds := GetAllFeeds()

	// Collect all recent items
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

	// Score all items
	type scoredItem struct {
		item  FeedItem
		score float64
	}

	var scored []scoredItem
	for _, item := range allItems {
		score := ScoreItem(&item, allItems)
		scored = append(scored, scoredItem{item, score})
	}

	// Sort by score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Build recommendations with diversity sampling
	recommendations := make([]RecommendedItem, 0) // Initialize as empty but non-nil slice
	targetCount := Cfg.ML.RecommendationCount
	diversityCount := (targetCount * Cfg.ML.DiversitySamplingPercent) / 100

	// Take top items
	topCount := targetCount - diversityCount
	for i := 0; i < topCount && i < len(scored); i++ {
		item := scored[i].item
		recommendations = append(recommendations, RecommendedItem{
			Title:  item.Title,
			Link:   item.Link,
			Age:    HumanizeAge(item.PublishedAt),
			Source: item.Source,
			Score:  scored[i].score,
			Reason: "High relevance",
		})
	}

	// Add some lower-scored items for diversity (avoid filter bubble)
	if len(scored) > topCount {
		remaining := scored[topCount:]
		// Shuffle remaining items and take some
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(remaining), func(i, j int) {
			remaining[i], remaining[j] = remaining[j], remaining[i]
		})

		for i := 0; i < diversityCount && i < len(remaining); i++ {
			item := remaining[i].item
			recommendations = append(recommendations, RecommendedItem{
				Title:  item.Title,
				Link:   item.Link,
				Age:    HumanizeAge(item.PublishedAt),
				Source: item.Source,
				Score:  remaining[i].score,
				Reason: "Diverse pick",
			})
		}
	}

	// Sort recommendations by score for final output
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	return recommendations
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

// HumanizeAge converts a timestamp to human-readable duration
func HumanizeAge(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "now"
	}
	if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", mins)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	}
	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1d ago"
	}
	return fmt.Sprintf("%dd ago", days)
}
