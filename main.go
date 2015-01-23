package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/blevesearch/bleve"
	bleveHttp "github.com/blevesearch/bleve/http"
)

var indexPath = flag.String("index", "fosdem.bleve", "index path")
var eventsPath = flag.String("events", "fosdem.ical", "fosdem events ical path")
var bindAddr = flag.String("addr", ":8099", "http listen address")

func main() {

	// turn on http request logging
	bleveHttp.SetLog(log.New(os.Stderr, "bleve.http", log.LstdFlags))

	// open index
	index, err := bleve.Open(*indexPath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		// or create new if it doesn't exist
		mapping := buildMapping()
		index, err = bleve.New(*indexPath, mapping)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer index.Close()

	// insert/update index in background
	go batchIndexEvents(index, *eventsPath)

	// start server
	startServer(index, *bindAddr)
}

func startServer(index bleve.Index, addr string) {
	// create a router to serve static files
	router := staticFileRouter()

	// add the API
	bleveHttp.RegisterIndexName("fosdem", index)
	searchHandler := bleveHttp.NewSearchHandler("fosdem")
	router.Handle("/api/search", searchHandler).Methods("POST")

	http.Handle("/", router)
	log.Printf("Listening on %v", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func batchIndexEvents(index bleve.Index, path string) {
	count := 0
	batch := bleve.NewBatch()
	for event := range parseEvents(path) {
		batch.Index(event.UID, event)
		if batch.Size() >= 100 {
			err := index.Batch(batch)
			if err != nil {
				log.Fatal(err)
			}
			count += batch.Size()
			log.Printf("Indexed %d Events\n", count)
			batch = bleve.NewBatch()
		}
	}
	if batch.Size() > 0 {
		err := index.Batch(batch)
		if err != nil {
			log.Fatal(err)
		}
		count += batch.Size()
	}
	log.Printf("Indexed %d Events\n", count)
}
