package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Finnhub API response structures
type FinnhubQuoteResponse struct {
	Symbol        string  `json:"symbol"`
	Price         float64 `json:"c"`
	PreviousClose float64 `json:"pc"`
	Change        float64 `json:"d"`
	ChangePercent float64 `json:"dp"`
}

// Fetch stock data from Finnhub
func fetchStockQuote(ctx context.Context, symbol string) (*StockData, error) {
	if Cfg.Stocks.API.APIKey == "" || Cfg.Stocks.API.APIKey == "YOUR_FINNHUB_API_KEY_HERE" {
		return nil, fmt.Errorf("finnhub API key not configured")
	}

	url := fmt.Sprintf("%s/quote?symbol=%s&token=%s",
		Cfg.Stocks.API.BaseURL, symbol, Cfg.Stocks.API.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var quote FinnhubQuoteResponse
	if err := json.Unmarshal(body, &quote); err != nil {
		return nil, err
	}

	if quote.Symbol == "" || quote.Price == 0 {
		fmt.Printf("DEBUG: No data in Finnhub response for %s. Full response: %s\n", symbol, string(body))
		return nil, fmt.Errorf("no data returned for symbol: %s", symbol)
	}

	return &StockData{
		Symbol:        symbol,
		Price:         quote.Price,
		Change:        quote.Change,
		ChangePercent: quote.ChangePercent,
		UpdatedAt:     time.Now(),
	}, nil
}

// Fetch currency exchange data
func fetchCurrencyExchange(ctx context.Context, fromCurrency, toCurrency string) (*StockData, error) {
	if Cfg.Stocks.API.APIKey == "" || Cfg.Stocks.API.APIKey == "YOUR_FINNHUB_API_KEY_HERE" {
		return nil, fmt.Errorf("finnhub API key not configured")
	}

	// Finnhub uses format like "EURUSD" for forex
	pair := fromCurrency + toCurrency
	url := fmt.Sprintf("%s/quote?symbol=%s&token=%s",
		Cfg.Stocks.API.BaseURL, pair, Cfg.Stocks.API.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var quote FinnhubQuoteResponse
	if err := json.Unmarshal(body, &quote); err != nil {
		return nil, err
	}

	if quote.Price == 0 {
		return nil, fmt.Errorf("no data returned for %s/%s", fromCurrency, toCurrency)
	}

	symbol := fmt.Sprintf("%s/%s", fromCurrency, toCurrency)
	return &StockData{
		Symbol:    symbol,
		Price:     quote.Price,
		UpdatedAt: time.Now(),
	}, nil
}

// Refresh all stock data
func refreshStockData(ctx context.Context) error {
	for _, stockCfg := range Cfg.Stocks.Items {
		if stockCfg.BaseCurrency != "" {
			// Currency exchange
			data, err := fetchCurrencyExchange(ctx, stockCfg.Symbol, stockCfg.BaseCurrency)
			if err != nil {
				fmt.Printf("Error fetching currency %s/%s: %v\n", stockCfg.Symbol, stockCfg.BaseCurrency, err)
				continue
			}
			data.Label = stockCfg.Label
			StockCacheMu.Lock()
			StockCache[stockCfg.Symbol] = data
			StockCacheMu.Unlock()
		} else {
			// Stock quote
			data, err := fetchStockQuote(ctx, stockCfg.Symbol)
			if err != nil {
				fmt.Printf("Error fetching stock %s: %v\n", stockCfg.Symbol, err)
				continue
			}
			data.Label = stockCfg.Label
			StockCacheMu.Lock()
			StockCache[stockCfg.Symbol] = data
			StockCacheMu.Unlock()
		}
	}
	return nil
}

// Start background worker for stock data refresh
func RefreshStockWorker(ctx context.Context) {
	if !Cfg.Stocks.Enabled {
		fmt.Println("Stock refresh disabled in config; skipping stock worker")
		return
	}

	// Initial fetch on startup
	refreshStockData(ctx)

	// Use configurable refresh interval from config (default: 3 hours for 4 stocks)
	// Finnhub free tier allows 60 API calls/minute, unlimited daily
	interval := time.Duration(Cfg.Stocks.Refresh) * time.Hour
	if interval < 1*time.Hour {
		interval = 1 * time.Hour // Enforce minimum 1-hour interval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refreshStockData(ctx)
		}
	}
}
