package application

type FeedIdList struct {
	Id     int
	Name   string
	Column int
}

func GetFeedIdList() ([]FeedIdList, error) {

	count, _ := appdb.Count()
	feeds, _ := appdb.GetAll(count)

	feedInfo := []FeedIdList{}

	for _, v := range feeds {
		feedInfo = append(feedInfo, FeedIdList{v.Id(), v.Name(), v.Column()})
	}

	return feedInfo, nil
}
