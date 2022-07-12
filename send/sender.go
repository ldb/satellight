package send

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ldb/satellight/protocol"
	"log"
	"net/http"
	"time"
)

const defaultRetries = 15

// Sender sends messages in a reliable way
type Sender struct {
	lastID int
	size   int
	id     int

	q chan Message

	currentMsg Message

	client    *http.Client
	endpoint  string
	retries   int
	nextRetry time.Time
}

func NewSender(id, queueSize int, endpoint string) *Sender {
	s := new(Sender)
	s.id = id
	s.q = make(chan Message, queueSize)
	s.endpoint = endpoint
	s.client = http.DefaultClient
	return s
}

// message is the internal representation of a message to send.
// It needs to export fields for JSON marshalling.
type message struct {
	ID        int                            `json:"id"`
	Timestamp time.Time                      `json:"ts"`
	Data      protocol.SpaceMessageMarshaler `json:"data"`
}

func (s *Sender) sendMessage(msg Message) error {
	m := message{
		ID:        msg.id,
		Timestamp: time.Now(),
		Data:      msg.Payload,
	}
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("error encoding protocol: %w", err)
	}
	buf := bytes.NewReader(b)
	req, err := http.NewRequest(http.MethodPost, s.endpoint, buf)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	res, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error received bad status: %v", err)
	}
	return nil
}

func backoff(retries int) time.Time {
	return time.Now().Add(time.Duration(retries) * time.Second)
}

func (s *Sender) Run() {
	for {
		m := s.currentMsg
		if m.id == 0 {
			m = <-s.q
			log.Printf("[%d] dequeued protocol %d", s.id, m.id)
		}

		if !s.nextRetry.Before(time.Now()) {
			time.Sleep(s.nextRetry.Sub(time.Now()))
		}

		if err := s.sendMessage(m); err != nil {
			s.retries += 1
			s.nextRetry = backoff(s.retries)
			s.currentMsg = m
			log.Printf("[%d]current protocol is %d, err: %v", s.id, m.id, err)
			continue
		}
		log.Printf("[%d] successfully sent protocol %d", s.id, m.id)
		s.currentMsg = Message{}
		s.retries = 0
		s.nextRetry = time.Now()
	}
}

func (s *Sender) EnqueueMessage(message Message) int {
	s.lastID += 1
	message.id = s.lastID
	s.q <- message
	return s.lastID
}

type Message struct {
	id      int
	Payload protocol.SpaceMessageMarshaler
}
