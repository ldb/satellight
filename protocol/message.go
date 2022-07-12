package protocol

import (
	"encoding/json"
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
	KindAdjustTime
)

type SpaceMessage struct {
	Kind Kind `json:"kind"`

	SenderID int

	Location   Location `json:"loc"`
	OzoneLevel float64  `json:"ol"`
	//Timestamp time.Time
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`
}

func (m *SpaceMessage) MarshalSpaceMessage() ([]byte, error) {
	return json.Marshal(m)
}

func (m *SpaceMessage) UnmarshalSpaceMessage(data []byte) error {
	return json.Unmarshal(data, m)
}
