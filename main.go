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
	// set up the commandline flags (-mongo and -pgurl)
	reqLog := flag.Bool("reqlog", false, "Enables request logging -- do NOT use in production")
	serverURL := flag.String("server", "localhost:3001", "The full URL for the root of the server")
	dbName := flag.String("dbname", "fhir", "Mongo database name")
	idxConfigPath := flag.String("idxconfig", "config/indexes.conf", "Path to the indexes config file")
	mongoHost := flag.String("mongohost", "localhost", "the hostname of the mongo database")
	readOnly := flag.Bool("readonly", false, "Run the API in read-only mode (no creates, updates, or deletes allowed)")
	debug := flag.Bool("debug", false, "Enables debug output for the mgo driver")
	disableCISearches := flag.Bool("disable-ci-searches", false, "Disables case-insensitive searches using regexes")
	dbTimeout := flag.String("db-timeout", "1m", "Database timeout, for example 45s, 1m, 300ms, etc.")

	flag.Parse()

	config := server.DefaultConfig

	// If these flags aren't set, the default values are used.
	config.ServerURL = *serverURL
	config.DatabaseName = *dbName
	config.IndexConfigPath = *idxConfigPath
	config.DatabaseHost = *mongoHost
	config.ReadOnly = *readOnly

	if *debug {
		mgo.SetDebug(true)
		var aLogger *log.Logger
		aLogger = log.New(os.Stderr, "", log.LstdFlags)
		mgo.SetLogger(aLogger)
	}

	config.EnableCISearches = !*disableCISearches
	duration, err := time.ParseDuration(*dbTimeout)
	if err == nil {
		config.DatabaseTimeout = duration
	}

	// setup the server
	s := server.NewServer(config)

	if *reqLog {
		s.Engine.Use(server.RequestLoggerHandler)
	}

	s.Run()
}
