package main

import (
	"fmt"
	"github.com/ldb/satellight/protocol"
	"github.com/ldb/satellight/receive"
	"github.com/ldb/satellight/send"
	"log"
	"math"
	"sync"
	"time"
)

const (
	considerOzoneCritical  = 0.4
	groundStationSenderID  = 0
	defaultSenderQueueSize = 20
)

type GroundStation struct {
	satellites map[int]*satellite
	receiver   *receive.Receiver
	mu         sync.RWMutex
}

type satellite struct {
	loc    protocol.Location
	sender *send.Sender
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
			log.Printf("message too old, ignoring")
			return
		}
		satellitedID := message.SenderID
		log.Printf("new message from satellite %d", satellitedID)
		g.mu.RLock()
		_, ok := g.satellites[satellitedID]
		g.mu.RUnlock()
		if !ok {
			g.mu.Lock()
			g.satellites[satellitedID] = &satellite{
				sender: send.NewSender(defaultSenderQueueSize, fmt.Sprintf("%s:%d", *satelliteAddress, satellitesBasePort+satellitedID)),
			}
			g.mu.Unlock()
		}
		g.mu.Lock()
		g.satellites[satellitedID].loc = message.Location
		g.mu.Unlock()

		if message.Kind != protocol.KindOzoneLevel {
			return
		}
		if message.OzoneLevel >= considerOzoneCritical {
			return
		}
		log.Printf("satellite %d found critical ozone levels!: %f", satellitedID, message.OzoneLevel)
		closestSatellite, distance := g.locateClosestSatellite(message.Location)
		log.Printf("satellite %d is closest to the zone (%.2fkm)", closestSatellite, distance)
		g.sendSatelliteToOzoneHole(closestSatellite, message.Location)
		log.Printf("sent satellite %d to fix the ozone hole", closestSatellite)
	}
}

// locateClosestSatellite calculates the distance of each satellite to
// loc and returns the ID of the satellite with the lowest distance.
func (g *GroundStation) locateClosestSatellite(loc protocol.Location) (int, float64) {
	dist := math.MaxFloat64
	minID := 0
	for id, sat := range g.satellites {
		distance := 1.609344 * 3963.0 * math.Acos((math.Sin(loc.Lat)*math.Sin(sat.loc.Lat))+math.Cos(loc.Lat)*math.Cos(sat.loc.Lat)*math.Cos(sat.loc.Lng-loc.Lng))
		if distance < dist {
			dist = distance
			minID = id
		}
	}
	return minID, dist
}

func (g *GroundStation) sendSatelliteToOzoneHole(id int, loc protocol.Location) {
	g.mu.RLock()
	sat, ok := g.satellites[id]
	g.mu.RUnlock()
	if !ok {
		return
	}
	sat.sender.EnqueueMessage(send.Message{
		Payload: &protocol.SpaceMessage{
			Kind:      protocol.KindAdjustCourse,
			Location:  loc,
			SenderID:  groundStationSenderID,
			Timestamp: time.Now(),
		},
	})
}

func (g *GroundStation) Run() error {
	return g.receiver.Run()
}
