package main

import (
	"flag"
	"log"

	"github.com/synthetichealth/gofhir/ptstats"

	"github.com/intervention-engine/fhir/server"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	// set up the commandline flags (-mongo and -pgurl)
	mongoHost := flag.String("mongohost", "localhost", "the hostname of the mongo database")
	pgURL := flag.String("pgurl", "", "The PG connection URL (e.g., postgres://username:password@localhost/dbname?sslmode=disable)")
	flag.Parse()

	if *pgURL == "" {
		log.Fatal("You must supply a pgurl flag value")
	}

	// configure the GORM Postgres driver and database connection
	db, err := gorm.Open("postgres", *pgURL)
	db.SingularTable(true) // disable table name pluralization globally

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// ping the db to ensure we connected successfully
	if err := db.DB().Ping(); err != nil {
		log.Fatal(err)
	}

	// setup the server
	s := server.NewServer(*mongoHost)

	// register interceptors
	s.AddInterceptor(op, resourceType, handler)
	s.AddInterceptor(op, resourceType, handler)

	// run the server
	s.Run(server.Config{
		UseSmartAuth:         false,
		UseLoggingMiddleware: false,
	})
}
