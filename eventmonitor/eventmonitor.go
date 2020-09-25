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

const (
	_interferenceType   = "RF_INT_DET"
	_powerType          = "TX_TYPE"
	_batteryCycleType   = "BATT_CYCLE"
	_batteryRunTimeType = "BATT_RUN_TIME"
	_batteryTypeType    = "BATT_TYPE"
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
	// Throw events for driver errors
	case driver.ErrorReportType:
		s.EventEmitter.Send(shure.Event{
			Key:    "Error",
			Value:  fmt.Sprintf("Error from driver: %s", r.FullReport),
			Device: dev,
		})
	case _batteryCycleType:
		// Trim proceeding 0's
		value := strings.TrimLeft(r.Value, "0")

		// Handle special values
		switch value {
		case "65535":
			value = "UNKNOWN"
		case "":
			value = "0"
		}

		s.EventEmitter.Send(shure.Event{
			Key:    "battery-cycles",
			Value:  value,
			Device: dev,
		})
	case _batteryRunTimeType:
		// Trim proceeding 0's
		value := strings.TrimLeft(r.Value, "0")

		// Handle special values
		switch value {
		case "65535":
			value = "UNKNOWN"
		case "65534":
			value = "CALCULATING"
		case "":
			value = "0"
		}

		s.EventEmitter.Send(shure.Event{
			Key:    "battery-charge-minutes",
			Value:  value,
			Device: dev,
		})
	case _batteryTypeType:
		// Handle special values
		value := r.Value
		switch r.Value {
		case "UNKN":
			value = "UNKNOWN"
		}

		s.EventEmitter.Send(shure.Event{
			Key:    "battery-type",
			Value:  value,
			Device: dev,
		})
	case _interferenceType:
		s.EventEmitter.Send(shure.Event{
			Key:    "interference",
			Value:  r.Value,
			Device: dev,
		})
	case _powerType:
		// Handle special values
		value := "ON" // On in any state other than UNKN
		if r.Value == "UNKN" {
			value = "STANDBY"
		}
		s.EventEmitter.Send(shure.Event{
			Key:    "power",
			Value:  value,
			Device: dev,
		})
	default:
		// skip
	}
}
