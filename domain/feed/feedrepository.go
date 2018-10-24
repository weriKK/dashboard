package feed

type FeedRepository interface {
	GetAll(limit int) ([]Feed, error)
	GetById(id int) (*Feed, error)
	Count() (int, error)
}
