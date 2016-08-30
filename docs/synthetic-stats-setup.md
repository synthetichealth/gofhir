Setup Synthetic Stats
=====================

Clone GoFHIR Server Code
------------------------
```
mkdir -p $GOPATH/src/github.com/synthetichealth
cd $GOPATH/src/github.com/synthetichealth
export https_proxy=http://gatekeeper.mitre.org:80
git clone https://github.com/synthetichealth/gofhir.git
unset https_proxy
```
You may need to install a missing C language dependency for Mgo, the golang Mongo package:

```
sudo apt-get install libsasl2-dev
```

Then install all project dependencies:

```
cd gofhir
go install ./...
```

Setup Postgres Test Database
----------------------------
You need to run the setup scripts as the `postgres` user:

```
sudo su - postgres
cd /home/<your_username>/go/src/github.com/synthetichealth/gofhir/postgres
./pgsetup.sh
logout
```

Test The Stats Package
----------------------
Using the test Postgres database (called `fhir_test`) test the stats package:

```
cd $GOPATH/src/github.com/synthetichealth/gofhir/stats
go test
```

Remove the Test Database
------------------------
After you're done testing, you can optionally remove the `fhir_test` database:

```
sudo su - postgres
cd /home/<your_username>/go/src/github.com/synthetichealth/gofhir/postgres
./pgcleanup.sh
```

Add The Disease and Fact Tables
-------------------------------

Again as user `postgres`, run `addfacts.sh` to create the following new tables:

- `synth_ma.synth_disease`
- `synth_ma.synth_county_facts`
- `synth_ma.synth_cousub_facts`

This creates the new disease table and populates it with diseases in `postgres/data/disease.csv`. Then every permutation of disease/county and disease/cousub is added to the county and cousub fact tables, respectively.

```
sudo su - postgres
cd /home/<your_username>/go/src/github.com/synthetichealth/gofhir/postgres
./addfacts.sh
```

Reset Synthetic Statistics
--------------------------
You can reset all of the synthetic statistics by running `resetstats.sh`.

```
sudo su - postgres
cd /home/<your_username>/go/src/github.com/synthetichealth/gofhir/postgres
./resetstats.sh
```