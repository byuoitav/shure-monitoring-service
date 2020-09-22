package shure

type Event struct {
	Device string
	Key    string
	Value  string
}

type EventEmitter interface {
	Send(Event)
}
