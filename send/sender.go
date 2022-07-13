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

// Sender sends messages in a reliable way by queueing them and retrying any failed delivery.
// Sender is designed to be used only by a single caller and thus not concurrency safe!
type Sender struct {
	// lastID tracks the ID of the last message that has entered the queue to provide consistent increments.
	lastID int

	queue chan message
	// currently dequeued Message for attempted delivery
	current message

	// This can be used to pass a custom logger to the sender. Left unset it defaults to the log default logger.
	Logger *log.Logger

	client   *http.Client
	endpoint string

	// Number of retries for delivery that have already been attempted
	retries   int
	nextRetry time.Time
}

func NewSender(queueSize int, endpoint string, logger *log.Logger) *Sender {
	s := new(Sender)
	s.Logger = logger
	s.queue = make(chan message, queueSize)
	s.endpoint = endpoint
	s.client = http.DefaultClient
	go s.run()
	return s
}

// message is the internal representation of a message to send.
// It needs to export fields for JSON marshalling.
type message struct {
	ID        int                            `json:"id"`
	Timestamp time.Time                      `json:"ts"`
	Data      protocol.SpaceMessageMarshaler `json:"data"`
}

func (s *Sender) sendMessage(msg message) error {
	msg.Timestamp = time.Now()
	b, err := json.Marshal(msg)
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

// run starts the main runloop for the sender. It dequeues any messages and attempts delivery.
func (s *Sender) run() {
	for {
		m := s.current
		if m.ID == 0 {
			m = <-s.queue
		}

		if !s.nextRetry.Before(time.Now()) {
			time.Sleep(s.nextRetry.Sub(time.Now()))
		}

		if err := s.sendMessage(m); err != nil {
			s.retries += 1
			s.nextRetry = backoff(s.retries)
			s.current = m
			s.Logger.Printf("message %d failed to deliver, retrying in %ds", m.ID, s.retries)
			continue
		}
		s.Logger.Printf("successfully sent message %d", m.ID)
		s.current = message{}
		s.retries = 0
		s.nextRetry = time.Now()
	}
}

func (s *Sender) EnqueueMessage(msg *protocol.SpaceMessage) int {
	s.lastID += 1
	m := message{
		Data:      msg,
		Timestamp: time.Now(),
		ID:        s.lastID,
	}
	s.queue <- m
	return s.lastID
}
