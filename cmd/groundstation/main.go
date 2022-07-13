package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"time"
)

const satellitesBasePort = 9000

var satelliteAddress = flag.String("satellites", "http://localhost", "Base URL of the satellites")
var groundStationAddress = flag.String("groundstation", ":8000", "address to listen on")

func main() {
	flag.Parse()

	// Seeding RNG.
	rand.Seed(time.Now().Unix())

	l := log.New(os.Stdout, "GS: ", log.Ltime)
	g := NewGroundStation(*groundStationAddress, l)
	log.Printf("groundstation started listening on %s", *groundStationAddress)
	close := make(chan struct{})
	_ = g.Run() // Ignoring the shutdown function.
	<-close
}
