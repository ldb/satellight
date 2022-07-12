package main

import (
	"fmt"
	"github.com/ldb/satellight/send"
	"log"
	"time"
)

func main() {
	sender := send.NewSender(5, "http://localhost:8000")
	go sender.Run()

	log.Println("started sender")

	i := 0
	for {
		i++
		sender.EnqueueMessage(send.Message{Payload: []byte(fmt.Sprintf("%d", i))})
		log.Printf("enqueued message %d", i)
		time.Sleep(1 * time.Second)
	}
}
