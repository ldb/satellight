package main

import (
	"flag"
	"log"
	"os"
)

const satellitesBasePort = 9000

var satelliteAddress = flag.String("satellites", "http://localhost", "Base URL of the satellites")
var groundStationAddress = flag.String("groundstation", ":8000", "address to listen on")

func main() {
	flag.Parse()
	l := log.New(os.Stdout, "GS: ", log.Ltime)
	g := NewGroundStation(*groundStationAddress, l)
	log.Printf("groundstation started listening on %s", *groundStationAddress)
	log.Fatal(g.Run())
}
