package main

import (
	"database/sql"
	"flag"
	"log"

	"github.com/intervention-engine/fhir/server"
	_ "github.com/lib/pq"
	"github.com/synthetichealth/gofhir/stats"
	"gopkg.in/mgo.v2"
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
	db, err := sql.Open("postgres", *pgURL)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// ping the db to ensure we connected successfully
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	da := stats.NewPgStatsDataAccess(db)

	// setup and run the server
	s := server.NewServer(*mongoHost)

	// Register patient interceptors
	s.AddInterceptor("Create", "Patient", &stats.PatientStatsCreateInterceptor{DataAccess: da})
	s.AddInterceptor("Update", "Patient", &stats.PatientStatsUpdateInterceptor{DataAccess: da})
	s.AddInterceptor("Delete", "Patient", &stats.PatientStatsDeleteInterceptor{DataAccess: da})

	// Register condition interceptors
	// The condition interceptors also require a mongodb connection
	session, err := mgo.Dial(*mongoHost)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	mdb := session.DB("fhir")
	mda := server.NewMongoDataAccessLayer(mdb, make(map[string]server.InterceptorList))
	s.AddInterceptor("Create", "Condition", &stats.ConditionStatsCreateInterceptor{PgDataAccess: da, MongoDataAccess: mda})
	s.AddInterceptor("Update", "Condition", &stats.ConditionStatsUpdateInterceptor{PgDataAccess: da, MongoDataAccess: mda})
	s.AddInterceptor("Delete", "Condition", &stats.ConditionStatsDeleteInterceptor{PgDataAccess: da, MongoDataAccess: mda})

	s.Run(server.Config{
		UseSmartAuth:         false,
		UseLoggingMiddleware: false,
	})
}
