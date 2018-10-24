package application

type FeedIdList struct {
	Id   int
	Name string
}

func GetFeedIdList() ([]FeedIdList, error) {

	count, _ := appdb.Count()
	feeds, _ := appdb.GetAll(count)

	feedInfo := []FeedIdList{}

	for _, v := range feeds {
		feedInfo = append(feedInfo, FeedIdList{v.Id(), v.Name()})
	}

	return feedInfo, nil
}
