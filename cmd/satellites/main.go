package main

import (
	"flag"
	"log"
	"time"
)

var satelliteCount = flag.Int("satelliteCount", 5, "Count of satellites launched")
var endpoint = flag.String("endpoint", "", "Not the default endpoint")

func main() {
	flag.Parse()

	log.Println("Satellites go space")

	for i := 0; i < *satelliteCount; i++ {
		go NewSatellite(i, *endpoint).Orbit()
		time.Sleep(5 * time.Second)
	}
}
