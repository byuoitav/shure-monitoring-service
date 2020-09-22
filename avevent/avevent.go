package avevent

import (
	"fmt"
	"time"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/shure-monitoring-service"
)

type Service struct {
	m *messenger.Messenger
}

func NewEmitter(hubAddress string) (*Service, error) {
	m, err := messenger.BuildMessenger(hubAddress, base.Messenger, 1000)
	if err != nil {
		return nil, fmt.Errorf("Error while trying to build messenger: %s", err)
	}

	return &Service{
		m: m,
	}, nil
}

func (s *Service) Send(e shure.Event) {
	devInfo := events.GenerateBasicDeviceInfo(e.Device)
	newEvent := events.Event{
		GeneratingSystem: "central-shure-monitoring",
		Timestamp:        time.Now(),
		TargetDevice:     devInfo,
		AffectedRoom:     devInfo.BasicRoomInfo,
		Key:              e.Key,
		Value:            e.Value,
	}

	s.m.SendEvent(newEvent)
}
