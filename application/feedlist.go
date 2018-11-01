package application

type FeedIdList struct {
	Id        int
	Name      string
	Url       string
	Rss       string
	Column    int
	ItemLimit int
}

func GetFeedIdList() ([]FeedIdList, error) {

	count, _ := appdb.Count()
	feeds, _ := appdb.GetAll(count)

	feedInfo := []FeedIdList{}

	for _, v := range feeds {
		feedInfo = append(feedInfo, FeedIdList{v.Id(), v.Name(), v.Url(), v.Rss(), v.Column(), v.Limit()})
	}

	return feedInfo, nil
}
