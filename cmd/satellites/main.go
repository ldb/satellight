package main

import (
	"flag"
	"log"
	"time"

	"github.com/ldb/satellight/protocol"
	"github.com/ldb/satellight/send"
)

var groundStationAddress = flag.String("groundstation", "http://localhost:8000", "URL of the ground station")

func main() {
	flag.Parse()

	log.Println("started sender")

	for i := 0; i < 5; i++ {
		go sendStuff(i)
	}
	sendStuff(6)
}

func sendStuff() {
	sender := send.NewSender(5, "http://localhost:8000")
	go sender.Run()
	i := 0
	for {
		i++
		sender.EnqueueMessage(send.Message{Payload: &protocol.SpaceMessage{Kind: protocol.KindAdjustTime, OzoneLevel: generateOzoneLevel()}})
		log.Printf("enqueued protocol %d", i)
		time.Sleep(1 * time.Second)
	}
}
