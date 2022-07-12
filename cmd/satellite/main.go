package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/ldb/satellight/protocol"
	"github.com/ldb/satellight/send"
)

func main() {

	log.Println("started sender")

	for i := 0; i < 5; i++ {
		go sendStuff(i)
	}
	sendStuff(6)
}

func sendStuff(id int) {
	sender := send.NewSender(id, 5, "http://localhost:8000")
	go sender.Run()
	i := 0
	for {
		i++
		sender.EnqueueMessage(send.Message{Payload: &protocol.SpaceMessage{Kind: protocol.KindAdjustTime, OzoneLevel: generateOzoneLevel()}})
		log.Printf("enqueued protocol %d", i)
		time.Sleep(1 * time.Second)
	}
}

func generateOzoneLevel() float64 {
	return rand.Float64()/2 + rand.Float64()/2
}
