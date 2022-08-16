package events

type Fetcher interface {
	Fetch(limit int) ([]Event, error)
}

type Processor interface {
	Process(e Event) error
}

type Type int

const (
	Unknown Type = iota // iota используется при объявлении групп констант. Первой константе присваивается значение нуль.
	Message
)

type Event struct {
	Type Type
	Text string
	Meta interface{}
}
