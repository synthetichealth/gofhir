package main

import (
	"flag"
	"log"
	"os"

	mgo "gopkg.in/mgo.v2"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	_ "github.com/synthetichealth/gofhir/synthma"
)

func main() {
	// set up the commandline flags (-mongo and -pgurl)
	reqLog := flag.Bool("reqlog", false, "Enables request logging -- do NOT use in production")
	serverURL := flag.String("server", "", "The full URL for the root of the server")
	dbName := flag.String("dbname", "fhir", "Mongo database name")
	idxConfigPath := flag.String("idxconfig", "config/indexes.conf", "Path to the indexes config file")
	mongoHost := flag.String("mongohost", "localhost", "the hostname of the mongo database")
	readOnly := flag.Bool("readonly", false, "Run the API in read-only mode (no creates, updates, or deletes allowed)")
	debug := flag.Bool("debug", false, "Enables debug output for the mgo driver")
	disableCISearches := flag.Bool("disable-ci-searches", false, "Disables case-insensitive searches using regexes")

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

	if *readOnly {
		s.Engine.Use(ReadOnlyHandler)
	}

	if *debug {
		mgo.SetDebug(true)
		var aLogger *log.Logger
		aLogger = log.New(os.Stderr, "", log.LstdFlags)
		mgo.SetLogger(aLogger)
	}

	if *disableCISearches {
		config.EnableCISearches = false
	}

	s.Run(config)
}

// ReadOnlyHandler makes the API read-only and responds to any requests that are not
// GET, HEAD, or OPTIONS with a 403 Forbidden error.
func ReadOnlyHandler(c *gin.Context) {

	method := c.Request.Method
	switch method {
	// allowed methods:
	case "GET", "HEAD", "OPTIONS":
		return
	// all other methods:
	default:
		c.AbortWithStatus(403)
	}
}
