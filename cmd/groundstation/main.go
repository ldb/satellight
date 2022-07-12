package main

import (
	"flag"
	"fmt"
	"github.com/ldb/satellight/protocol"
	"github.com/ldb/satellight/receive"
	"log"
)

var satelliteAddress = flag.String("satellites", "http://localhost:9000", "Base URL of the satellites")
var groundStationAddress = flag.String("groundstation", ":8000", "address to listen on")

func main() {
	flag.Parse()

	r := receive.NewReceiver(*groundStationAddress, func(message protocol.SpaceMessage) {
		fmt.Printf("%+v", message)
	})
	log.Fatal(r.Run())

}
