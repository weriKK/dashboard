package feed

type Format int

type Feed struct {
	id               int
	name             string
	url              string
	rss              string
	column           int
	visibleItemLimit int
}

func New(id int, name string, url string, rss string, column int, visibleItemLimit int) Feed {
	return Feed{id, name, url, rss, column, visibleItemLimit}
}

func (f Feed) Id() int      { return f.id }
func (f Feed) Name() string { return f.name }
func (f Feed) Url() string  { return f.url }
func (f Feed) Rss() string  { return f.rss }
func (f Feed) Column() int  { return f.column }
func (f Feed) Limit() int   { return f.visibleItemLimit }
