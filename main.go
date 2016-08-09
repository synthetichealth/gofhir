package main

import (
	"flag"
	"log"

	"gitlab.mitre.org/synthea/gofhir/ptstats"

	"github.com/intervention-engine/fhir/server"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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
	db.SingularTable(true)  // disable table name pluralization globally

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// ping the db to ensure we connected successfully
	if err := db.DB().Ping(); err != nil {
		log.Fatal(err)
	}
	// configure the stat interceptor
	ptStatsInterceptor := &ptstats.PtStatsInterceptor{
		CousubDA: ptstats.PgCountySubdivisionDataAccess{DB: db},        
		SynthCountyStatsDA: ptstats.PgSyntheticCountyStatsDataAccess{DB: db},
		SynthCousubStatsDA: ptstats.PgSyntheticCountySubdivisionStatsDataAccess{DB: db},
	}

	// setup and run the server
	s := server.NewServer(*mongoHost)
	s.Engine.Use(ptStatsInterceptor.Handler)
	s.Run(server.Config{
		UseSmartAuth:         false,
		UseLoggingMiddleware: false,
	})
}
