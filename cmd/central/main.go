package main

import (
	"log"
	"sync"

	"github.com/byuoitav/shure-monitoring-service"
	"github.com/byuoitav/shure-monitoring-service/avevent"
	"github.com/byuoitav/shure-monitoring-service/couch"
	"github.com/byuoitav/shure-monitoring-service/eventmonitor"
	"github.com/spf13/pflag"
)

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
	for _, r := range recvs {
		go func(r shure.Receiver) {
			_ = m.Monitor(r)
		}(r)
	}

	log.Printf("Monitoring initialized on %d receivers", len(recvs))

	// Hang forever
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
