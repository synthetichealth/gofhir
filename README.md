FHIR Server w/ Interceptor
===================================================================================================================================================================

This project builds on the [Go-based FHIR server](https://github.com/intervention-engine/ie) by providing a simple interceptor that is invoked on all requests.  It's currently only stubbed out, but the intent is for the interceptor to extracts stats and store them in a PG database.

Building and Running the Server Locally
---------------------------------

This project works standalone -- in that although it is built on the Go-based FHIR server, that server is already embedded in this project.

For information on installing and running only the FHIR server, please begin by referencing the following sections of the IE guide:

-	(Prerequisite) [Install Git](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-git)
-	(Prerequisite) [Install Go](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-go)
-	(Prerequisite) [Install MongoDB](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#install-mongodb)
-	(Prerequisite) [Run MongoDB](https://github.com/intervention-engine/ie/blob/master/docs/dev_install.md#run-mongodb)

Following standard Go practices, you should clone the *fhir* repository under your `$GOPATH` src folder, using a package-based sub-path:

```
$ mkdir -p $GOPATH/src/gitlab.mitre.org/synthea
$ cd $GOPATH/src/gitlab.mitre.org/synthea
$ git clone https://gitlab.mitre.org/synthea/gofhir.git
```

Before you can run the FHIR server, you should build the `gofhir` executable:

```
$ cd $GOPATH/src/gitlab.mitre.org/synthea/gofhir
$ go build
```

The above commands do not need to be run again unless you make (or download) changes to the *gofhir* source code.

Once the executable is built, you can run it with the `-pgurl` argument:

```
$ ./gofhir -pgurl postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full
```

The *gofhir* server accepts connections on port 3001 by default.

License
-------

Copyright 2016 The MITRE Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
