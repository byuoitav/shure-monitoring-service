package avevent

import (
	"fmt"
	"log"
	"time"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/shure-monitoring-service"
)

type LogEventEmitter struct {
	m *messenger.Messenger
}

func NewLogEmitter(hubAddress string) (*LogEventEmitter, error) {
	m, err := messenger.BuildMessenger(hubAddress, base.Messenger, 1000)
	if err != nil {
		return nil, fmt.Errorf("Error while trying to build messenger: %s", err)
	}

	return &LogEventEmitter{
		m: m,
	}, nil
}

func (e *LogEventEmitter) Send(event shure.Event) {
	// Log first
	log.Printf("Event: Key: %s | Value: %s | Device: %s", event.Key, event.Value, event.Device)

	// Emit event to av central hub
	devInfo := events.GenerateBasicDeviceInfo(event.Device)
	newEvent := events.Event{
		GeneratingSystem: "central-shure-monitoring",
		Timestamp:        time.Now(),
		TargetDevice:     devInfo,
		AffectedRoom:     devInfo.BasicRoomInfo,
		Key:              event.Key,
		Value:            event.Value,
	}

	e.m.SendEvent(newEvent)
}
