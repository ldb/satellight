package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

var satelliteAddress = flag.String("satellites", "http://localhost:9000", "Base URL of the satellites")

func main() {
	flag.Parse()

	log.Fatal(http.ListenAndServe(":8000", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		log.Printf("received: %q", string(b))
	})))
}
