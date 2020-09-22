package eventmonitor

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	driver "github.com/byuoitav/shure-driver"
	"github.com/byuoitav/shure-monitoring-service"
)

type Service struct {
	EventEmitter shure.EventEmitter
}

func (s *Service) Monitor(r shure.Receiver) error {

	var d *driver.ULXDReceiver
	var err error

	// Keep retrying every 5 minutes if we fail to create a new driver
	for d == nil {
		d, err = driver.NewReceiver(r.Address)
		if err != nil {
			s.EventEmitter.Send(shure.Event{
				Key:    "Error",
				Value:  fmt.Sprintf("Error while initializing driver: %s", err),
				Device: r.Name,
			})
			time.Sleep(5 * time.Minute)
		}
	}

	// Eternal loop
	for {
		c, err := d.StartReporting(context.TODO())
		if err != nil {
			s.EventEmitter.Send(shure.Event{
				Key:    "Error",
				Value:  fmt.Sprintf("Error while initializing reporting: %s", err),
				Device: r.Name,
			})
			// wait 5 minutes after a failed connection and retry
			time.Sleep(5 * time.Minute)
			continue
		}

		// Debug purposes. Delete later
		log.Printf("Started Reporting for %s", r.Name)

		// Range over reports until the channel is closed
		for report := range c {
			s.processReport(report, r)
		}
	}
}

func (s *Service) processReport(r driver.Report, recv shure.Receiver) {

	// Default device to reciever
	dev := recv.Name

	// Try to make the name match the Microphone if name is what we expect
	// and the channel is not 0 (all channels) or -1 (error)
	pieces := strings.Split(recv.Name, "-")
	if len(pieces) >= 3 && r.Channel > 0 {
		dev = fmt.Sprintf("%s-%s-MIC%d", pieces[0], pieces[1], r.Channel)
	}
	// Dispatch by type
	switch r.Type {
	case driver.ERROR:
		// Skip unknown reports
		if r.Value != "UnrecognizedReport" {
			s.EventEmitter.Send(shure.Event{
				Key:    "Error",
				Value:  fmt.Sprintf("Error from driver: %s", r.Message),
				Device: dev,
			})
		}
	case driver.BATTERY_CYCLES:
		s.EventEmitter.Send(shure.Event{
			Key:    "battery-cycles",
			Value:  r.Value,
			Device: dev,
		})
	case driver.BATTERY_CHARGE_MINUTES:
		s.EventEmitter.Send(shure.Event{
			Key:    "battery-charge-minutes",
			Value:  r.Value,
			Device: dev,
		})
	case driver.BATTERY_TYPE:
		s.EventEmitter.Send(shure.Event{
			Key:    "battery-type",
			Value:  r.Value,
			Device: dev,
		})
	case driver.INTERFERENCE:
		s.EventEmitter.Send(shure.Event{
			Key:    "interference",
			Value:  r.Value,
			Device: dev,
		})
	case driver.POWER:
		s.EventEmitter.Send(shure.Event{
			Key:    "power",
			Value:  r.Value,
			Device: dev,
		})
	default:
		// skip
	}
}
