package receive

import (
	"encoding/json"
	"github.com/ldb/satellight/protocol"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const defaultListenAddress = ":8000"

type SpaceMessageHandler func(message protocol.SpaceMessage)

type Receiver struct {
	server http.Server
}

func NewReceiver(addr string, handle SpaceMessageHandler) *Receiver {
	r := Receiver{
		server: http.Server{
			Addr:    addr,
			Handler: handleMessage(handle),
		},
	}
	return &r
}

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

func (r *Receiver) Run() error {
	return r.server.ListenAndServe()
}
