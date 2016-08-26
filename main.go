package main

import (
	"flag"
	"log"

	"github.com/intervention-engine/fhir/server"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/synthetichealth/gofhir/stats"
)

func main() {
	// set up the commandline flags (-mongo and -pgurl)
	mongoHost := flag.String("mongohost", "localhost", "the hostname of the mongo database")
	pgURL := flag.String("pgurl", "", "The PG connection URL (e.g., postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full)")
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
	da := stats.NewPgStatsDataAccess(db)

	// setup and run the server
	s := server.NewServer(*mongoHost)

	// register interceptors
	s.AddInterceptor("Create", "Patient", &stats.PatientStatsCreateInterceptor{DataAccess: da})
	s.AddInterceptor("Update", "Patient", &stats.PatientStatsUpdateInterceptor{DataAccess: da})
	s.AddInterceptor("Delete", "Patient", &stats.PatientStatsDeleteInterceptor{DataAccess: da})

	s.Run(server.Config{
		UseSmartAuth:         false,
		UseLoggingMiddleware: false,
	})
}
