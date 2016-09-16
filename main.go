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
	reqLog := flag.Bool("reqlog", false, "Enables request logging -- do NOT use in production")
	dbName := flag.String("dbname", "fhir", "Mongo database name")
	idxConfigPath := flag.String("idxconfig", "config/indexes.conf", "Path to the indexes config file")
	mongoHost := flag.String("mongohost", "localhost", "the hostname of the mongo database")
	pgURL := flag.String("pgurl", "", "The PG connection URL (e.g., postgres://fhir:fhir@localhost/fhir?sslmode=disable)")

	flag.Parse()

	// setup the server
	s := server.NewServer(*mongoHost)

	config := server.DefaultConfig

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

	// configure the Postgres driver and database connection
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

	// Register patient interceptors
	s.AddInterceptor("Create", "Patient", stats.NewPatientStatsCreateInterceptor(da))
	s.AddInterceptor("Delete", "Patient", stats.NewPatientStatsDeleteInterceptor(da))

	// Register condition interceptors
	// The condition interceptors also require a mongodb connection
	session, err := mgo.Dial(*mongoHost)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	mdb := session.DB("fhir")
	mda := server.NewMongoDataAccessLayer(mdb, make(map[string]server.InterceptorList))
	s.AddInterceptor("Create", "Condition", stats.NewConditionStatsCreateInterceptor(da, mda))
	s.AddInterceptor("Update", "Condition", stats.NewConditionStatsUpdateInterceptor(da, mda))
	s.AddInterceptor("Delete", "Condition", stats.NewConditionStatsDeleteInterceptor(da, mda))

	s.Run(config)
}
