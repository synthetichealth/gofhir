package main

import (
	"flag"
	"log"

	"github.com/intervention-engine/fhir/server"
	_ "github.com/lib/pq"
)

func main() {
	// set up the commandline flags (-mongo and -pgurl)
	reqLog := flag.Bool("reqlog", false, "Enables request logging -- do NOT use in production")
	serverURL := flag.String("server", "", "The full URL for the root of the server")
	dbName := flag.String("dbname", "fhir", "Mongo database name")
	idxConfigPath := flag.String("idxconfig", "config/indexes.conf", "Path to the indexes config file")
	mongoHost := flag.String("mongohost", "localhost", "the hostname of the mongo database")
	pgURL := flag.String("pgurl", "", "The PG connection URL (e.g., postgres://fhir:fhir@localhost/fhir?sslmode=disable)")

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

	if *pgURL == "" {
		log.Fatal("You must supply a pgurl flag value")
	}

	s.Run(config)
}
