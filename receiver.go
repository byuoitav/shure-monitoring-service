package shure

type Receiver struct {
	Name    string
	Address string
}

type ReceiverStore interface {
	GetReceivers() ([]Receiver, error)
}
