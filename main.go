package main

import (
	"bytes"
	"database/sql"
	"flag"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	_ "github.com/lib/pq"
)

// StatExtractor is the middleware that handles the stat extraction
type StatExtractor struct {
	db *sql.DB
}

// Handler is invoked on every request to the Go FHIR server
func (s *StatExtractor) Handler(c *gin.Context) {
	// We probably only care about posts and puts...
	if c.Request != nil && (c.Request.Method == "POST" || c.Request.Method == "PUT") {
		// Read the body and close the stream
		buf, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()

		// Do something with it -- replace the line below to use the sql package to do whatever
		log.Printf("body: %s", buf)

		// We need to replenish the body since we drained the stream
		c.Request.Body = ioutil.NopCloser(bytes.NewReader(buf))
	}
	// Go to the next handler
	c.Next()
}

func main() {
	// set up the commandline flags (-mongo and -pgurl)
	mongoHost := flag.String("mongohost", "localhost", "the hostname of the mongo database")
	pgURL := flag.String("pgurl", "", "The PG connection URL (e.g., postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full)")
	flag.Parse()

	if *pgURL == "" {
		log.Fatal("You must supply a pgurl flag value")
	}

	// setup the pg driver connection
	db, err := sql.Open("postgres", *pgURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// ping the db to ensure we connected successfully
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	// configure the stat extractor
	statExtractor := &StatExtractor{db: db}

	// setup and run the server
	s := server.NewServer(*mongoHost)
	s.Engine.Use(statExtractor.Handler)
	s.Run(server.Config{
		UseSmartAuth:         false,
		UseLoggingMiddleware: false,
	})
}
