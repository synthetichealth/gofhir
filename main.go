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
	serverURL := flag.String("server", server.DefaultConfig.ServerURL, "The full URL for the root of the server")
	idxConfigPath := flag.String("idxconfig", server.DefaultConfig.IndexConfigPath, "Path to the indexes config file")
	readOnly := flag.Bool("readonly", server.DefaultConfig.ReadOnly, "Run the API in read-only mode (no creates, updates, or deletes allowed)")
	debug := flag.Bool("debug", server.DefaultConfig.Debug, "Enables debug level logging")

	// database options
	mongoHost := flag.String("db.host", server.DefaultConfig.DatabaseHost, "the hostname of the mongo database")
	dbName := flag.String("db.name", server.DefaultConfig.DatabaseName, "Mongo database name")
	noCountResults := flag.Bool("db.no-count-results", !server.DefaultConfig.CountTotalResults, "Stops searches from counting the total results, saving time")
	disableCISearches := flag.Bool("db.disable-ci-searches", !server.DefaultConfig.EnableCISearches, "Disables case-insensitive searches using regexes")
	dbSocketTimeout := flag.String("db.socket-timeout", server.DefaultConfig.DatabaseSocketTimeout.String(), "Database socket timeout, for example 45s, 1m, 300ms, etc.")
	dbOpTimeout := flag.String("db.op-timeout", server.DefaultConfig.DatabaseOpTimeout.String(), "Database opereation timeout, for example 45s, 1m, 300ms, etc.")

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
