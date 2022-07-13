package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

var satelliteCount = flag.Int("satelliteCount", 5, "Count of satellites launched")
var endpoint = flag.String("endpoint", "http://localhost:8000", "Groundstation endpoint")

const standartPort = 9000

func main() {
	flag.Parse()

	// Seeding RNG.
	rand.Seed(time.Now().Unix())

	log.Println("Satellites go space")

	// Spin up new satellites.
	wg := sync.WaitGroup{}
	for i := 1; i <= *satelliteCount; i++ {
		wg.Add(1)
		go func(i int) {
			l := log.New(os.Stdout, fmt.Sprintf("[%d]: ", i), log.Ltime)
			err := NewSatellite(i, *endpoint, l).Orbit()
			if err != nil {
				l.Printf("error received from satellite: %v", err)
				wg.Done()
				return
			}
		}(i)
	}
	wg.Wait()
}
