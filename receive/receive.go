package receive

import (
	"context"
	"encoding/json"
	"github.com/ldb/satellight/protocol"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Callback handler for received messages.
type SpaceMessageHandler func(message protocol.SpaceMessage)

type Receiver struct {
	server http.Server
	logger *log.Logger
}

// NewReceiver creates a new receiver.
func NewReceiver(addr string, handle SpaceMessageHandler, log *log.Logger) *Receiver {
	r := Receiver{
		server: http.Server{
			Addr:    addr,
			Handler: handleMessage(handle),
		},
		logger: log,
	}
	return &r
}

// handleMessage handles an incoming message by unmarshalling it from JSON and passing it into the
// provided SpaceMessageHandler callback function.
func handleMessage(msgHandle SpaceMessageHandler) http.HandlerFunc {
	type message struct {
		ID        int                   `json:"id"`
		Timestamp time.Time             `json:"ts"`
		Data      protocol.SpaceMessage `json:"data"`
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			log.Printf("error reading request: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		msg := new(message)
		if err := json.Unmarshal(buf, &msg); err != nil {
			log.Printf("error unmarshaling SpaceMessage: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		msgHandle(msg.Data)
		writer.WriteHeader(http.StatusOK)
	}
}

// Run is the main runloop for the receiver.
// It concurrently runs an HTTP server and returns a function to stop the receiver.
func (r *Receiver) Run() func() {
	go func() {
		r.logger.Println("starting receiver")
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.logger.Println("error running receiver: %v", err)
		}
	}()
	return func() {
		r.logger.Println("stopping receiver")
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		if err := r.server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
			r.logger.Printf("error stopping receiver: %v", err)
		}
	}
}
