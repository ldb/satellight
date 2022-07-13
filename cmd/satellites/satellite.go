package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ldb/satellight/protocol"
	"github.com/ldb/satellight/receive"
	"github.com/ldb/satellight/send"
)

const defaultQueueSize = 5
const defaultEndpoint = "http://localhost:8000"

type Satellite struct {
	ID     int
	sender *send.Sender

	Logger *log.Logger

	CurrentLocation  protocol.Location
	TargetLocation   protocol.Location
	currentlySteered bool
	ts               time.Time
}

// NewSatellite initializes a new satellite.
func NewSatellite(id int, endpoint string, logger *log.Logger) *Satellite {
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	initalLoc := initialLocation()
	return &Satellite{
		ID:              id,
		sender:          send.NewSender(defaultQueueSize, endpoint, logger),
		Logger:          logger,
		CurrentLocation: initalLoc,
		TargetLocation:  initalLoc,
		ts:              time.Now(),
	}
}

// ReadOzoneLevel reads the Ozone Levels at the current location.
// The probability of values around 0,5 is higher than towards the edges.
func (s *Satellite) ReadOzoneLevel() float64 {
	return rand.Float64()/2 + rand.Float64()/2
}

// Random initial location. Satellites are launched into Orbit from all over the world.
func initialLocation() protocol.Location {
	return protocol.Location{
		Lat: rand.Float64()*(90-(-90)) - 90,   // Range for Latitude [-90,90)
		Lng: rand.Float64()*(180-(-180)) - 90, // Range for Longitude [-180,180)
		Alt: rand.Float64()*1800 + 160,        // Range for Altitude [160,1960)
	}
}

// Very naïve implementation of a limited update, we might leave the cosmic sphere that way :o.
func (s *Satellite) nextLocation() protocol.Location {
	return protocol.Location{
		Lat: s.CurrentLocation.Lat + rand.Float64()*0.1,
		Lng: s.CurrentLocation.Lng + rand.Float64()*0.1,
		Alt: s.CurrentLocation.Alt + rand.Float64()*0.1,
	}
}

// Steer moves the satellite by a distance towards a new (random) location.
// This function should only be invoked when the satellite was instructed to adjust it's location by the ground-station.
func (s *Satellite) Steer(location protocol.Location) {
	s.TargetLocation = location
}

// Orbit is the main run-loop of a satellite.
func (s *Satellite) Orbit() error {
	// Handle messages received by the satellite asynchronously.
	handler := receive.SpaceMessageHandler(func(message protocol.SpaceMessage) {
		switch kind := message.Kind; kind {
		case protocol.KindAdjustCourse:
			s.Steer(message.Location)
			s.currentlySteered = true
			s.Logger.Printf("Groundstation steered us to new location %.2fkm away: %+v",
				s.CurrentLocation.Distance(message.Location),
				s.CurrentLocation)
		default:
			s.Logger.Printf("Received message from ground-station of kind %d", message.Kind)
			break
		}
	})

	receiver := receive.NewReceiver(fmt.Sprintf(":%d", standartPort+s.ID), handler, s.Logger)

	// receiver runs concurrently to the rest of this function.
	stopReceiver := receiver.Run()
	s.Logger.Println("going to space huiiiii")
	time.Sleep(time.Second)
	s.Logger.Println("reached space, starting to orbit")

	for {
		// If the ground-station is not steering us, randomly move around in space.
		if !s.currentlySteered {
			s.TargetLocation = s.nextLocation()
		}

		// The satellite is lost to the ground station
		if rand.Float64() < 0.01 {
			stopReceiver()
			return errors.New("My battery is low and it's getting dark :(")
		}

		// Send messages with current ozone level to groundstation
		currentLevel := s.ReadOzoneLevel()
		s.sender.EnqueueMessage(&protocol.SpaceMessage{
			SenderID:   s.ID,
			Kind:       protocol.KindOzoneLevel,
			Timestamp:  time.Now(),
			OzoneLevel: currentLevel,
			Location:   s.CurrentLocation,
		})
		
		distance := s.CurrentLocation.Distance(s.TargetLocation)
		s.Logger.Printf("flying to new location %.2fkm away", distance)
		time.Sleep(time.Duration(distance) * time.Second / 1000)
		s.CurrentLocation = s.TargetLocation
		s.currentlySteered = false
	}
}
