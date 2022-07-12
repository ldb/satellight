package main

import (
	"github.com/ldb/satellight/protocol"
	"github.com/ldb/satellight/receive"
	"github.com/ldb/satellight/send"
	"log"
	"sync"
	"time"
)

type GroundStation struct {
	satellites map[int]*satellite
	receiver   *receive.Receiver
	mu         sync.Mutex
}

type satellite struct {
	currentLocation protocol.Location
	sender          *send.Sender
}

func NewGroundStation(addr string) *GroundStation {
	g := &GroundStation{
		satellites: make(map[int]*satellite),
	}
	g.receiver = receive.NewReceiver(addr, g.handle())
	return g
}

func (g *GroundStation) handle() receive.SpaceMessageHandler {
	return func(message protocol.SpaceMessage) {
		if message.Kind == protocol.KindInvalid {
			return
		}
		if time.Now().Sub(message.Timestamp) > 20*time.Second {
			log.Printf("ignoring old message")
			return
		}
		satellitedID := message.SenderID
		g.mu.Lock()
		g.satellites[satellitedID].currentLocation = message.Location
		g.mu.Unlock()

		if message.Kind != protocol.KindOzoneLevel {

		}

	}
}

func (g *GroundStation) Run() error {
	return g.receiver.Run()
}
