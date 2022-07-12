package send

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"
)

const defaultRetries = 15

// Sender sends messages in a reliable way
type Sender struct {
	lastID int
	size   int

	q chan Message

	currentMsg Message

	client    *http.Client
	endpoint  string
	retries   int
	nextRetry time.Time
}

func NewSender(queueSize int, endpoint string) *Sender {
	s := new(Sender)
	s.q = make(chan Message, queueSize)
	s.endpoint = endpoint
	s.client = http.DefaultClient
	return s
}

func (s *Sender) sendMessage(message Message) error {
	b := bytes.NewReader(message.Payload)
	req, err := http.NewRequest(http.MethodPost, s.endpoint, b)
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
			log.Printf("dequeued message %d", m.id)
		}

		if !s.nextRetry.Before(time.Now()) {
			time.Sleep(s.nextRetry.Sub(time.Now()))
		}

		if err := s.sendMessage(m); err != nil {
			s.retries += 1
			s.nextRetry = backoff(s.retries)
			s.currentMsg = m
			log.Printf("current message is %d, err: %v", m.id, err)
			continue
		}
		log.Printf("successfully sent message %d", m.id)
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
	Payload []byte
}
