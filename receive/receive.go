package receive

import "net/http"

const defaultListenAddress = ":8000"

type Receiver struct {
	server http.Server
}
