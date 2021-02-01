package main

import (
	"log"
	"time"

	"github.com/byuoitav/shure-monitoring-service"
	"github.com/byuoitav/shure-monitoring-service/avevent"
	"github.com/byuoitav/shure-monitoring-service/couch"
	"github.com/byuoitav/shure-monitoring-service/eventmonitor"
	"github.com/spf13/pflag"
)

// _onlineEventInterval specifies how often online events are sent for the receiver
var _onlineEventInterval = 3 * time.Minute

// _onlineThreshold specifies how long a reciever must be online to be considered online
var _onlineThreshold = 30 * time.Second

func main() {
	var (
		dbAddr       string
		dbUser       string
		dbPass       string
		eventHubAddr string
	)

	pflag.StringVar(&dbAddr, "db-address", "", "The address to the couch database")
	pflag.StringVar(&dbUser, "db-username", "", "The username for the couch database")
	pflag.StringVar(&dbPass, "db-password", "", "The password for the couch database")
	pflag.StringVar(&eventHubAddr, "eventhub-address", "", "The address for the event hub")

	pflag.Parse()

	c, err := couch.New(dbAddr, dbUser, dbPass)
	if err != nil {
		log.Panicf("Failed to initialize couch: %s", err)
	}

	recvs, err := c.GetReceivers()
	if err != nil {
		log.Panicf("Failed to get receivers from database: %s", err)
	}

	e, err := avevent.NewLogEmitter(eventHubAddr, "central-shure-monitoring")
	if err != nil {
		log.Panicf("Failed to start log emitter")
	}

	m := eventmonitor.Service{
		EventEmitter: e,
	}

	log.Printf("Beginning monitoring...")

	// Initialize monitoring
	for i := range recvs {
		go func(r *shure.Receiver) {
			_ = m.Monitor(r)
		}(&recvs[i])
	}

	log.Printf("Monitoring initialized on %d receivers", len(recvs))

	// Wait 1 minute before reporting the first time
	// This allows time to register the receivers as online
	time.Sleep(1 * time.Minute)

	// report online status for each receiver every 3 minutes
	for {
		for i := range recvs {
			recvs[i].OnlineMu.RLock()
			// If the receiver has been online for more than 30 seconds then report it as online
			if recvs[i].Online && time.Since(recvs[i].LastUpdated) > _onlineThreshold {
				e.Send(shure.Event{
					Key:    "online",
					Value:  "Online",
					Device: recvs[i].Name,
				})
			} else {
				e.Send(shure.Event{
					Key:    "online",
					Value:  "Offline",
					Device: recvs[i].Name,
				})
			}
			recvs[i].OnlineMu.RUnlock()
		}
		time.Sleep(_onlineEventInterval)
	}
}
