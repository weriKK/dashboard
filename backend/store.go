package backend

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	db            *sql.DB
	TokenWeights  map[string]float64
	TokenWeightMu sync.RWMutex
)

// OpenStore opens the SQLite database and creates tables if needed
func OpenStore(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// WAL mode for better concurrent read performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("failed to set WAL mode: %w", err)
	}

	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

func createTables() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS click_events (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			item_key   TEXT NOT NULL,
			title      TEXT NOT NULL,
			link       TEXT NOT NULL,
			source     TEXT NOT NULL,
			category   TEXT NOT NULL,
			clicked_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS token_weights (
			token      TEXT PRIMARY KEY,
			weight     REAL NOT NULL,
			updated_at DATETIME NOT NULL
		);
	`)
	return err
}

// SaveClickEvent persists a single click event and updates token weights
func SaveClickEvent(feedback ClickFeedback) error {
	_, err := db.Exec(
		`INSERT INTO click_events (item_key, title, link, source, category, clicked_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		feedback.ItemKey, feedback.ItemTitle, feedback.ItemLink,
		feedback.Source, feedback.Category, feedback.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to save click event: %w", err)
	}

	tokens := Tokenize(feedback.ItemTitle)
	if len(tokens) == 0 {
		return nil
	}

	// Each token gets an equal share of the click weight
	weightPerToken := Cfg.ML.ClickWeight / float64(len(tokens))

	TokenWeightMu.Lock()
	defer TokenWeightMu.Unlock()

	for _, token := range tokens {
		TokenWeights[token] += weightPerToken

		_, err := db.Exec(
			`INSERT INTO token_weights (token, weight, updated_at) VALUES (?, ?, ?)
			 ON CONFLICT(token) DO UPDATE SET weight = ?, updated_at = ?`,
			token, TokenWeights[token], time.Now(),
			TokenWeights[token], time.Now(),
		)
		if err != nil {
			log.Printf("Failed to save token weight for %q: %v", token, err)
		}
	}

	return nil
}

// LoadTokenWeights reads all token weights from the database into memory
func LoadTokenWeights() error {
	TokenWeightMu.Lock()
	defer TokenWeightMu.Unlock()

	TokenWeights = make(map[string]float64)

	rows, err := db.Query("SELECT token, weight FROM token_weights")
	if err != nil {
		return fmt.Errorf("failed to load token weights: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var token string
		var weight float64
		if err := rows.Scan(&token, &weight); err != nil {
			return fmt.Errorf("failed to scan token weight: %w", err)
		}
		TokenWeights[token] = weight
		count++
	}

	log.Printf("Loaded %d token weights from database", count)
	return rows.Err()
}

// ApplyTokenDecay multiplies all token weights by the decay factor and persists them.
// Call this once per scoring pass (e.g. on each dashboard request is too frequent;
// a daily or hourly schedule is better — but for simplicity we run it at startup).
func ApplyTokenDecay() error {
	decay := Cfg.ML.TokenDecayPerDay
	if decay <= 0 || decay >= 1 {
		return nil
	}

	TokenWeightMu.Lock()
	defer TokenWeightMu.Unlock()

	now := time.Now()
	rows, err := db.Query("SELECT token, weight, updated_at FROM token_weights")
	if err != nil {
		return fmt.Errorf("failed to read token weights for decay: %w", err)
	}
	defer rows.Close()

	type entry struct {
		token     string
		weight    float64
		updatedAt time.Time
	}
	var entries []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.token, &e.weight, &e.updatedAt); err != nil {
			return err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, e := range entries {
		daysSinceUpdate := now.Sub(e.updatedAt).Hours() / 24
		if daysSinceUpdate < 0.5 {
			continue // skip very recent entries
		}
		decayed := e.weight * math.Pow(decay, daysSinceUpdate)
		TokenWeights[e.token] = decayed

		if _, err := tx.Exec(
			"UPDATE token_weights SET weight = ?, updated_at = ? WHERE token = ?",
			decayed, now, e.token,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// PruneOldEvents removes click events older than the retention window
func PruneOldEvents(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result, err := db.Exec("DELETE FROM click_events WHERE clicked_at < ?", cutoff)
	if err != nil {
		return fmt.Errorf("failed to prune old events: %w", err)
	}
	deleted, _ := result.RowsAffected()
	if deleted > 0 {
		log.Printf("Pruned %d click events older than %d days", deleted, retentionDays)
	}
	return nil
}

// CloseStore closes the database connection
func CloseStore() {
	if db != nil {
		db.Close()
	}
}
