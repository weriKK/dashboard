package application

import (
	"dashboard/internal/domain/feed"
)

var appdb feed.FeedRepository

func Init(fr feed.FeedRepository) {
	appdb = fr
}
