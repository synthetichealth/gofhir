FHIR Server w/ Interceptor
===================================================================================================================================================================

This project builds on the [Go-based FHIR server](https://github.com/intervention-engine/ie) by providing data-layer interceptors to track patient statistics. These statistics are stored in a Postgres database.

Building and Running the Server Locally
---------------------------------

This project works standalone -- in that although it is built on the Go-based FHIR server, that server is already embedded in this project.

For information on installing and running only the FHIR server, please begin by referencing the following sections of the IE guide:

-	(Prerequisite) [Install Git](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-git)
-	(Prerequisite) [Install Go](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-go)
-	(Prerequisite) [Install MongoDB](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-mongodb)
-	(Prerequisite) [Run MongoDB](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#run-mongodb)
-  (Prerequisite) [Install Postgres and Extensions](https://github.com/synthetichealth/gofhir/blob/stats/postgres/postgres-setup.md#install-postgres) **needs updated link after merge with master
-  (Prerequisite) [Run Postgres](https://github.com/synthetichealth/gofhir/blob/stats/postgres/postgres-setup.md#run-postgres) **needs updated link after merge with master

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

For development and testing you should setup a local Postgres database. Follow these [Setup Development Database](https://github.com/synthetichealth/gofhir/blob/ptstats/postgres/postgres-setup.md#setup-development-database) instructions.

Once the executable is built, you can run it with the `-pgurl` argument:

```
$ ./gofhir -pgurl postgres://fhir_test:fhir_test@localhost/fhir_test?sslmode=disable
```

The *gofhir* server accepts connections on port 3001 by default.

License
-------

Copyright 2016 The MITRE Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
