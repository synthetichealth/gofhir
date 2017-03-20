package main

import (
	"flag"
	"log"
	"os"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/intervention-engine/fhir/server"
	_ "github.com/synthetichealth/gofhir/synthma"
)

func main() {
	// server options
	reqLog := flag.Bool("reqlog", false, "Enables request logging -- do NOT use in production")
	serverURL := flag.String("server", "localhost:3001", "The full URL for the root of the server")
	idxConfigPath := flag.String("idxconfig", "config/indexes.conf", "Path to the indexes config file")
	readOnly := flag.Bool("readonly", false, "Run the API in read-only mode (no creates, updates, or deletes allowed)")
	debug := flag.Bool("debug", false, "Enables debug level logging")

	// database options
	mongoHost := flag.String("db.host", "localhost", "the hostname of the mongo database")
	dbName := flag.String("db.name", "fhir", "Mongo database name")
	noCountResults := flag.Bool("db.no-count-results", false, "Stops searches from counting the total results, saving time")
	disableCISearches := flag.Bool("db.disable-ci-searches", false, "Disables case-insensitive searches using regexes")
	dbSocketTimeout := flag.String("db.socket-timeout", "1m", "Database socket timeout, for example 45s, 1m, 300ms, etc.")
	dbOpTimeout := flag.String("db.op-timeout", "1m", "Database opereation timeout, for example 45s, 1m, 300ms, etc.")

	flag.Parse()

	config := server.DefaultConfig

	// If these flags aren't set, the default values are used.
	config.ServerURL = *serverURL
	config.IndexConfigPath = *idxConfigPath
	config.ReadOnly = *readOnly
	config.DatabaseHost = *mongoHost
	config.DatabaseName = *dbName
	config.CountTotalResults = !*noCountResults
	config.EnableCISearches = !*disableCISearches

	if *debug {
		mgo.SetDebug(true)
		var aLogger *log.Logger
		aLogger = log.New(os.Stderr, "", log.LstdFlags)
		mgo.SetLogger(aLogger)
	}

	socketDuration, err := time.ParseDuration(*dbSocketTimeout)
	if err == nil {
		config.DatabaseSocketTimeout = socketDuration
	}

	opDuration, err := time.ParseDuration(*dbOpTimeout)
	if err == nil {
		config.DatabaseOpTimeout = opDuration
	}

	// setup the server
	s := server.NewServer(config)

	if *reqLog {
		s.Engine.Use(server.RequestLoggerHandler)
	}

	s.Run()
}
