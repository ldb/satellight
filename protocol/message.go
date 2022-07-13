package protocol

import (
	"encoding/json"
	"math"
	"time"
)

// The kind of controll message that can be exchanged between satellites and the ground-station.
type Kind int

const (
	KindInvalid      Kind = iota
	KindOzoneLevel        // Messages of this kind contain an ozone reading for the current location.
	KindAdjustCourse      // Messages of this kind contain a new target location of a satellite.
)

// Messages are really just marshalled into JSON when transmitted over the network.
type SpaceMessageMarshaler interface {
	MarshalSpaceMessage() ([]byte, error)
}

type SpaceMessageUnmarshaler interface {
	UnmarshalSpaceMessage([]byte) error
}

type SpaceMessage struct {
	Kind Kind `json:"kind"`

	SenderID int

	Location   Location `json:"loc"`
	OzoneLevel float64  `json:"ol"`
	Timestamp  time.Time
}

func (m *SpaceMessage) MarshalSpaceMessage() ([]byte, error) {
	return json.Marshal(m)
}

func (m *SpaceMessage) UnmarshalSpaceMessage(data []byte) error {
	return json.Unmarshal(data, m)
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`
}

// Distance calculates the distance between two locations in km.
// Altitude is not taken into account.
func (l *Location) Distance(loc Location) float64 {
	return 1.609344 * 3963.0 * math.Acos((math.Sin(loc.Lat)*math.Sin(l.Lat))+math.Cos(loc.Lat)*math.Cos(l.Lat)*math.Cos(l.Lng-loc.Lng))
}
