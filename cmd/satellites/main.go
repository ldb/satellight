package main

import (
	"flag"
	"log"
	"sync"
	"time"
)

var satelliteCount = flag.Int("satelliteCount", 5, "Count of satellites launched")
var endpoint = flag.String("endpoint", "http://localhost:8000", "Groundstation endpoint")

const standartPort = 9000

func main() {
	flag.Parse()

	log.Println("Satellites go space")

	wg := sync.WaitGroup{}
	for i := 0; i < *satelliteCount; i++ {
		wg.Add(1)
		go func() {
			err := NewSatellite(i, *endpoint).Orbit()
			if err != nil {
				log.Printf(err.Error())
				return
			}
		}()
		time.Sleep(5 * time.Second)
	}
	wg.Wait()
}
