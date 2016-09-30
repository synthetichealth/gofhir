package main

import (
	"flag"

	"github.com/intervention-engine/fhir/server"
)

func main() {
	// set up the commandline flags (-mongo and -pgurl)
	reqLog := flag.Bool("reqlog", false, "Enables request logging -- do NOT use in production")
	serverURL := flag.String("server", "", "The full URL for the root of the server")
	dbName := flag.String("dbname", "fhir", "Mongo database name")
	idxConfigPath := flag.String("idxconfig", "config/indexes.conf", "Path to the indexes config file")
	mongoHost := flag.String("mongohost", "localhost", "the hostname of the mongo database")

	flag.Parse()

	// setup the server
	s := server.NewServer(*mongoHost)

	config := server.DefaultConfig

	if *serverURL != "" {
		config.ServerURL = *serverURL
	}

	if *dbName != "" {
		config.DatabaseName = *dbName
	}

	if *idxConfigPath != "" {
		config.IndexConfigPath = *idxConfigPath
	}

	if *reqLog {
		s.Engine.Use(server.RequestLoggerHandler)
	}

	s.Run(config)
}
