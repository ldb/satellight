package main

import (
	"fmt"
	"github.com/ldb/satellight/protocol"
	"github.com/ldb/satellight/receive"
	"github.com/ldb/satellight/send"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

const (
	considerOzoneCritical  = 0.4
	groundStationSenderID  = 0
	defaultSenderQueueSize = 20
)

type GroundStation struct {
	logger     *log.Logger
	satellites map[int]*satellite
	receiver   *receive.Receiver
	mu         sync.RWMutex
}

type satellite struct {
	loc    protocol.Location
	sender *send.Sender
}

func NewGroundStation(addr string, logger *log.Logger) *GroundStation {
	g := &GroundStation{
		logger:     logger,
		satellites: make(map[int]*satellite),
	}
	g.receiver = receive.NewReceiver(addr, g.handle(), logger)
	return g
}

func (g *GroundStation) handle() receive.SpaceMessageHandler {
	return func(message protocol.SpaceMessage) {
		if message.Kind == protocol.KindInvalid {
			g.logger.Println("message of invalid kind")
			return
		}
		if message.SenderID == groundStationSenderID {
			g.logger.Println("message from invalid sender")
			return
		}
		td := time.Now().Sub(message.Timestamp)
		if td > 20*time.Second {
			g.logger.Printf("message too old, ignoring (%.2fs)", td.Seconds())
			return
		}
		satellitedID := message.SenderID
		g.logger.Printf("new message from satellite %d", satellitedID)
		g.mu.RLock()
		_, ok := g.satellites[satellitedID]
		g.mu.RUnlock()
		if !ok {
			sl := log.New(os.Stdout, fmt.Sprintf("SAT [%d] :", satellitedID), log.Ltime)
			g.mu.Lock()
			g.satellites[satellitedID] = &satellite{
				sender: send.NewSender(
					defaultSenderQueueSize,
					fmt.Sprintf("%s:%d", *satelliteAddress, satellitesBasePort+satellitedID),
					sl,
				),
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
		g.logger.Printf("satellite %d detected critical ozone levels!: %f", satellitedID, message.OzoneLevel)
		closestSatellite, distance := g.locateClosestSatellite(satellitedID, message.Location)
		g.logger.Printf("satellite %d is closest to the zone (%.2fkm)", closestSatellite, distance)
		g.sendSatelliteToOzoneHole(closestSatellite, message.Location)
		g.logger.Printf("sent satellite %d to fix the ozone hole", closestSatellite)
	}
}

// locateClosestSatellite calculates the distance of each satellite to
// loc and returns the ID of the satellite with the lowest distance.
func (g *GroundStation) locateClosestSatellite(exc int, loc protocol.Location) (int, float64) {
	dist := math.MaxFloat64
	minID := 0
	g.mu.RLock()
	for id, sat := range g.satellites {
		if id == exc {
			continue
		}
		distance := sat.loc.Distance(loc)
		if distance < dist {
			dist = distance
			minID = id
		}
	}
	g.mu.RUnlock()
	return minID, dist
}

func (g *GroundStation) sendSatelliteToOzoneHole(id int, loc protocol.Location) {
	g.mu.RLock()
	sat, ok := g.satellites[id]
	g.mu.RUnlock()
	if !ok {
		return
	}
	sat.sender.EnqueueMessage(&protocol.SpaceMessage{
		Kind:      protocol.KindAdjustCourse,
		Location:  loc,
		SenderID:  groundStationSenderID,
		Timestamp: time.Now(),
	})
}

func (g *GroundStation) Run() func() {
	return g.receiver.Run()
}
