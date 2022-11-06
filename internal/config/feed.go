package config

type FeedConfig []FeedItem

type FeedItem struct {
	Title     string `yaml:"title"`
	WebLink   string `yaml:"webLink"`
	FeedLink  string `yaml:"feedLink"`
	ColumnId  int    `yaml:"columnId"`
	ItemLimit int    `yaml:"itemLimit"`
}

func GetFeedConfig() FeedConfig {
	return cfg.FeedConfig
}
