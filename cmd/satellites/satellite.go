package main

import (
	"math/rand"
	"time"

	"github.com/ldb/satellight/protocol"
	"github.com/ldb/satellight/send"
)

const defaultQueueSize = 5
const defaultEndpoint = "http://localhost:8000"

type Satellite struct {
	ID     int
	sender *send.Sender

	Loc protocol.Location
	ts  time.Time
}

func NewSatellite(id int) *Satellite {
	return &Satellite{
		ID:     id,
		sender: send.NewSender(defaultQueueSize, defaultEndpoint),
		Loc:    randomLocation(),
		ts:     time.Now(),
	}
}

// ReadOzoneLevel reads the Ozone Levels at the current location.
func (s *Satellite) ReadOzoneLevel() float64 {
	return rand.Float64()/2 + rand.Float64()/2
}

func randomLocation() protocol.Location {
	return protocol.Location{
		Lat: rand.Float64()*(90-(-90)) - 90,   // Range for Latitude [-90,90)
		Lng: rand.Float64()*(180-(-180)) - 90, // Range for Longitude [-180,180)
		Alt: rand.Float64()*1800 + 160,        // Range for Altitude [160,1960)
	}
}

// Steer moves the satellite by a distance towards a new (random) location.
func (s *Satellite) Steer(distance float64) {
	s.Loc.Alt += distance
	s.Loc.Lat += distance
	s.Loc.Lng += distance
}

func (s *Satellite) Orbit() error {
	// Start receiver async
	// receive message
	// handle message based on kind
	//
	for {
		currentLevel := s.ReadOzoneLevel()
		s.sender.EnqueueMessage(send.Message{Payload: &protocol.SpaceMessage{
			Kind:       protocol.KindOzoneLevel,
			OzoneLevel: currentLevel,
			Location:   s.Loc,
		}})
	}
}
