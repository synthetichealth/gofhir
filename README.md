Synthetic Mass GoFHIR Server [![Build Status](https://travis-ci.org/synthetichealth/gofhir.svg?branch=master)](https://travis-ci.org/synthetichealth/gofhir)
============================

This project builds on the [Go-based FHIR server](https://github.com/intervention-engine/ie) by providing data-layer interceptors to track patient statistics. These statistics are stored in a Postgres database and used by the Synthetic Mass UI.

Building the Server Locally
---------------------------

This project works standalone -- in that although it is built on the Go-based FHIR server, that server is already embedded in this project.

For information on installing and running only the FHIR server, please begin by referencing the following sections of the IE guide:

-	(Prerequisite) [Install Git](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-git)
-	(Prerequisite) [Install Go](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-go)
-	(Prerequisite) [Install MongoDB](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-mongodb)
-	(Prerequisite) [Run MongoDB](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#run-mongodb)
-  (Prerequisite) [Install Postgres and Extensions](https://github.com/synthetichealth/gofhir/blob/master/docs/postgres-setup.md#install-postgres)
-  (Prerequisite) [Run Postgres](https://github.com/synthetichealth/gofhir/blob/master/docs/postgres-setup.md#run-postgres)

Following standard Go practices, you should clone the *fhir* repository under your `$GOPATH` src folder, using a package-based sub-path:

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


Running the Server Locally
--------------------------

For development and testing you should setup a local Postgres database. Follow these [Setup Development Database](https://github.com/synthetichealth/gofhir/blob/ptstats/postgres/postgres-setup.md#setup-development-database) instructions.

Once the executable is built, you can run it with the `-pgurl` argument:

```
$ ./gofhir -pgurl postgres://fhir:fhir@localhost/fhir?sslmode=disable
```

The *gofhir* server accepts connections on port 3001 by default.

Running the Server in Production
--------------------------------
In production you should make sure the following are set:

### Gin Mode

```
export GIN_MODE=release
```
This silences the debug logging.

### Server URL

Run GoFHIR with the `-server` flag to indicate the full URL for the root of the server. This is especially important when running GoFHIR behind a proxy; GoFHIR depends on the `ServerURL` configuration to build the correct pagination URLs when returning resource bundles.

License
-------

Copyright 2016 The MITRE Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
