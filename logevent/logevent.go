package logevent

import (
	"log"

	"github.com/byuoitav/shure-monitoring-service"
)

type Service struct {
}

func (s *Service) Send(e shure.Event) {
	log.Printf("Event: Key: %s | Value: %s | Device: %s", e.Key, e.Value, e.Device)
}
