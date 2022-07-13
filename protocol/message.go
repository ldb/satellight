package protocol

import (
	"encoding/json"
	"math"
	"time"
)

type Kind int

type SpaceMessageMarshaler interface {
	MarshalSpaceMessage() ([]byte, error)
}

type SpaceMessageUnmarshaler interface {
	UnmarshalSpaceMessage([]byte) error
}

const (
	KindInvalid Kind = iota
	KindOzoneLevel
	KindAdjustCourse
)

type SpaceMessage struct {
	Kind Kind `json:"kind"`

	SenderID int

	Location   Location `json:"loc"`
	OzoneLevel float64  `json:"ol"`
	Timestamp  time.Time
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`
}

// Distance calculates the distance between two locations in km.
func (l *Location) Distance(loc Location) float64 {
	return 1.609344 * 3963.0 * math.Acos((math.Sin(loc.Lat)*math.Sin(l.Lat))+math.Cos(loc.Lat)*math.Cos(l.Lat)*math.Cos(l.Lng-loc.Lng))
}

func (m *SpaceMessage) MarshalSpaceMessage() ([]byte, error) {
	return json.Marshal(m)
}

func (m *SpaceMessage) UnmarshalSpaceMessage(data []byte) error {
	return json.Unmarshal(data, m)
}
