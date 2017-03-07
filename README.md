Synthetic Mass GoFHIR Server [![Build Status](https://travis-ci.org/synthetichealth/gofhir.svg?branch=master)](https://travis-ci.org/synthetichealth/gofhir)
============================

This project builds on the [Go-based FHIR server](https://github.com/intervention-engine/fhir) by providing data-layer interceptors to track patient statistics. These statistics are stored in a Postgres database and used by the Synthetic Mass UI. Additional, custom indexes are also added over the base indexes that come with the FHIR server. The FHIR server supports FHIR STU3 1.8 (San Antonio), dated January 2017.

Building the Server Locally
---------------------------

This project works standalone -- in that although it is built on the Go-based FHIR server, that server is already embedded in this project.

For information on installing and running only the FHIR server, please begin by referencing the following sections of the IE guide:

-	(Prerequisite) [Install Git](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-git)
-	(Prerequisite) [Install Go](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-go)
-	(Prerequisite) [Install MongoDB](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-mongodb)
-	(Prerequisite) [Run MongoDB](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#run-mongodb)

Following standard Go practices, you should clone the *gofhir* repository under your `$GOPATH` src folder, using a package-based sub-path:

```
$ mkdir -p $GOPATH/src/github.com/synthetichealth
$ cd $GOPATH/src/github.com/synthetichealth
$ git clone https://github.com/synthetichealth/gofhir.git
```

Before you can run the FHIR server, you should build the `gofhir` executable:

```
$ cd $GOPATH/src/github.com/synthetichealth/gofhir
$ go build
```

The above commands do not need to be run again unless you make (or download) changes to the *gofhir* source code.

Options
-------
To get a list of options, run:

```
$ ./gofhir --help
```

Currently, the GoFHIR server supports the following options:

```
Usage of ./gofhir:
  -db-timeout string
    	Database timeout, for example 45s, 1m, 300ms, etc. (default "1m")
  -dbname string
    	Mongo database name (default "fhir")
  -debug
    	Enables debug output for the mgo driver
  -disable-ci-searches
    	Disables case-insensitive searches using regexes
  -idxconfig string
    	Path to the indexes config file (default "config/indexes.conf")
  -mongohost string
    	the hostname of the mongo database (default "localhost")
  -no-count-results
    	Stops searches from counting the total results, saving time
  -readonly
    	Run the API in read-only mode (no creates, updates, or deletes allowed)
  -reqlog
    	Enables request logging -- do NOT use in production
  -server string
    	The full URL for the root of the server (default "localhost:3001")
```

Running the Server Locally
--------------------------

You will need mongodb 3.2 or later running locally. To start the server locally run:

```
$ ./gofhir [OPTIONS]
```

Most of the default options should work out-of-the-box in a local development environment. The GoFHIR server accepts connections on port 3001 by default.

Running the Server in Production
--------------------------------
In production you should make sure the following are set:

### Gin Mode

```
export GIN_MODE=release
```
This silences Gin's debug logging.

### Server Options

**Do not use the `-debug` or `-reqlog` flags in production.**

### Server URL

Run GoFHIR with the `-server` flag to indicate the full URL for the root of the server. This is especially important when running GoFHIR behind a proxy; GoFHIR depends on the `ServerURL` configuration to build the correct pagination URLs when returning resource bundles.

License
-------

Copyright 2016 The MITRE Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
