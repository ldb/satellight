package main

import (
	"github.com/ldb/satellight/protocol"
	"github.com/ldb/satellight/send"
	"log"
	"time"
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
		sender.EnqueueMessage(send.Message{Payload: &protocol.SpaceMessage{Kind: protocol.KindAdjustTime, OzoneLevel: 9.9}})
		log.Printf("enqueued protocol %d", i)
		time.Sleep(1 * time.Second)
	}
}
