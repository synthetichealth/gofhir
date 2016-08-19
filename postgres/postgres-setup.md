Install and Run Postgres for Development
========================================

This project uses a local Postgres database for development and testing. This database contains the minimal schema, tables, and data needed to support the patient statistics interceptors. In production the server should be pointed at a production Postgres database containing the full suite of tables, schema, and data.

Install Postgres
----------------

Follow the [official installation instructions](https://www.postgresql.org/download/) for your operating system. At the time of this writing the latest Postgres version was 9.5.3.

Many Mac OSX developers use [homebrew](http://brew.sh/):

```
$ brew install postgres
```

Install PostGIS Extensions
--------------------------
The Synthetic Mass project uses the [PostGIS](http://postgis.net/) Postgres extension to add support for geographic objects. At the time of this writing the project used PostGIS 2.2.2. There are several extensions in the PostGIS package, but Synthetic Mass uses:

- **postgis** - enables PostGIS
- **fuzzystrmatch** - fuzzy matching needed for Tiger
- **postgis\_tiger\_geocoder** - enables U.S. Tiger Geocoder

On Mac, use homebrew:

```
$ brew install postgis
```

Run Postgres
------------
On Mac you can run postgres using `brew services`:

```
$ brew services start postgres
```

This starts Postgres with Mac OSX's `launchctl` manager. If you don't have [homebrew services](https://github.com/Homebrew/homebrew-services) installed run:

```
$ brew tap homebrew/services
```

To verify that Postgres is running, from the command line try:

```
$ postgres --version
```

Setup Development Database
--------------------------
With Postgres running, from the `postgres/` folder run `pgsetup.sh`:

```
$ cd $GOPATH/src/github.com/synthetichealth/gofhir/postgres/
$ ./pgsetup.sh
```

You may need to make `pgsetup.sh` an executable by running:

```
$ chmod +x pgsetup.sh
```

This creates a local database `fhir_test` with a `fhir_test` user, adds the needed PostGIS extensions, then sets up the tables for Synthetic Mass statistics. Once the database and tables are setup the script copies the needed statistical data from CSV files in the `postgres/data/` folder.

The following tables are relevant for Synthetic Mass statistics:

- `synth_ma.synth_county_stats` - synthetic patient statistics by county
- `synth_ma.synth_cousub_stats` - synthetic patient statistics by county subdivison (city/town)
- `tiger.cousub` - county subdivision information

All of this data is **for Massachusetts only**.

Connect to the Development Database
-----------------------------------
When you run the `gofhir` server specify the following `-pgurl`:

```
$ ./gofhir -pgurl postgres://fhir_test:fhir_test@localhost/fhir_test?sslmode=disable
```

This points the server to the local development database.

Removing the Development Database
---------------------------------
To remove the development database, from the `postgres/` folder run `pgcleanup.sh`:

```
$ cd $GOPATH/src/github.com/synthetichealth/gofhir/postgres/
$ ./pgcleanup.sh
```
You may need to make `pgcleanup.sh` an executable by running:

```
$ chmod +x pgcleanup.sh
```

This removes the `fhir_test` database and `fhir_test` user.
