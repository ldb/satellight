package send

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const defaultRetries = 15

// Sender sends messages in a reliable way
type Sender struct {
	queue     *queue
	buffer    map[int]Message
	maxBuffer int

	client      *http.Client
	endpoint    string
	retriesLeft int
	nextRetry   time.Time
}

func NewSender(queueSize, bufferSize int, endpoint string) *Sender {
	s := new(Sender)
	s.buffer = make(map[int]Message)
	s.maxBuffer = bufferSize
	s.queue = &queue{
		q: make(chan Message, queueSize),
	}
	s.endpoint = endpoint
	s.client = http.DefaultClient
	return s
}

func (s *Sender) sendMessage(message Message) error {
	if s.retriesLeft <= 0 {
		return errors.New("no retries left")
	}
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
	log.Printf("successfully sent message %d", message.id)
	s.retriesLeft = defaultRetries
	return nil
}

func backoff(retries int) time.Time {
	return time.Now().Add(time.Duration(defaultRetries-retries) * time.Second)
}

func (s *Sender) decrRetries() {
	if s.retriesLeft <= 0 {
		return
	}
	s.retriesLeft -= 1
}

func (s *Sender) Run() {
	for {
		for mid, m := range s.buffer {
			if s.retriesLeft > 0 || s.nextRetry.Before(time.Now()) {
				if err := s.sendMessage(m); err != nil {
					log.Printf("error sending message %d: %v", m.id, err)
					s.decrRetries()
					s.nextRetry = backoff(s.retriesLeft)
					log.Printf("retries left: %d, next retry: %s", s.retriesLeft, time.Now().Sub(s.nextRetry).String())
					s.buffer[mid] = m
					continue
				}
				delete(s.buffer, mid)
			}
		}

		if l := len(s.buffer); l >= s.maxBuffer {
			continue
		}
		m := s.queue.dequeue()

		if err := s.sendMessage(m); err != nil {
			s.decrRetries()
			s.nextRetry = backoff(s.retriesLeft)
			s.buffer[m.id] = m
		}
	}
}

func (s *Sender) SendMessage(message Message) int {
	return s.queue.enqueue(message)
}

type queue struct {
	lastID int
	size   int

	q chan Message
}

func (q *queue) enqueue(message Message) int {
	q.lastID += 1
	message.id = q.lastID
	q.q <- message
	return q.lastID
}

func (q *queue) dequeue() Message {
	return <-q.q
}

type Message struct {
	id      int
	Payload []byte
}
