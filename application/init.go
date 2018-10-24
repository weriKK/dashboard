package application

import (
	"github.com/weriKK/dashboard/domain/feed"
)

var appdb feed.FeedRepository

func Init(fr feed.FeedRepository) {
	appdb = fr
}
