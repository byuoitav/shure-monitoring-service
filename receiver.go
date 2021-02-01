package shure

import (
	"sync"
	"time"
)

type Receiver struct {
	Name    string
	Address string

	OnlineMu    sync.RWMutex
	Online      bool
	LastUpdated time.Time
}

type ReceiverStore interface {
	GetReceivers() ([]Receiver, error)
}
