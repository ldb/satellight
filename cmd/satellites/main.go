package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	for i := 1; i <= *satelliteCount; i++ {
		wg.Add(1)
		go func(i int) {
			l := log.New(os.Stdout, fmt.Sprintf("[%d]: ", i), log.Ltime)
			err := NewSatellite(i, *endpoint, l).Orbit()
			if err != nil {
				l.Printf(err.Error())
				return
			}
		}(i)
		time.Sleep(5 * time.Second)
	}
	wg.Wait()
}
