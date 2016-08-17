#!/bin/bash
# Sets up a local postgres database that supports SyntheticMass statistics.
# This creates the minimal schema and tables needed to support development and
# testing, so use it for development and testing only.
# 
# Carlton Duffett
# 08-15-2016

echo "Creating fhir_test database..."

PWD=`pwd`

# reset the local fhir database (if one exists)
psql <<EOF
SET client_encoding = 'UTF8';

-- Start fresh
DROP DATABASE IF EXISTS fhir_test;
DROP ROLE IF EXISTS fhir_test;
CREATE DATABASE fhir_test;
EOF

# Create the new schema and tables
psql -d fhir_test <<EOF
-- Create the database and adds PostGIS extensions
CREATE EXTENSION postgis;
CREATE EXTENSION fuzzystrmatch;
CREATE EXTENSION postgis_tiger_geocoder;

-- Create roles
CREATE USER fhir_test WITH PASSWORD 'fhir_test';

-- Create schema
CREATE SCHEMA synth_ma AUTHORIZATION fhir_test;
-- NOTE: The "tiger" schema is automatically created by PostGIS
GRANT USAGE ON SCHEMA tiger TO public;
GRANT SELECT ON TABLE tiger.cousub TO public;

-- Synthetic county statistics
CREATE TABLE synth_ma.synth_county_stats (
    ct_name character varying(100),
    ct_fips character varying(3) NOT NULL,
    sq_mi double precision,
    pop numeric,
    pop_male numeric,
    pop_female numeric,
    pop_sm double precision,
    ct_poly public.geometry(MultiPolygon,4269),
    ct_pnt public.geometry
);

ALTER TABLE synth_ma.synth_county_stats OWNER TO fhir_test;

-- Synthetic county subdivison statistics
CREATE TABLE synth_ma.synth_cousub_stats (
    ct_name character varying(100),
    ct_fips character varying(3) NOT NULL,
    cs_name character varying(100),
    cs_fips character varying(5) NOT NULL,
    sq_mi double precision,
    pop numeric,
    pop_male numeric,
    pop_female numeric,
    pop_sm double precision,
    cs_poly public.geometry(MultiPolygon,4269),
    cs_pnt public.geometry
);

ALTER TABLE synth_ma.synth_cousub_stats OWNER TO fhir_test;
EOF

# Copy data into the statistics tables
cat $PWD/data/synth_county.csv | psql -d fhir_test -c "\COPY synth_ma.synth_county_stats FROM STDIN (DELIMITER ',', QUOTE '\"', FORMAT CSV)"
cat $PWD/data/synth_cousub.csv | psql -d fhir_test -c "\COPY synth_ma.synth_cousub_stats FROM STDIN (DELIMITER ',', QUOTE '\"', FORMAT CSV)"
cat $PWD/data/cousub.csv | psql -d fhir_test -c "\COPY tiger.cousub FROM STDIN (DELIMITER ',', QUOTE '\"', FORMAT CSV)"

# cleanup after database creation and population
psql -d fhir_test <<EOF
VACUUM ANALYZE synth_ma.synth_county_stats;
VACUUM ANALYZE synth_ma.synth_cousub_stats;
VACUUM ANALYZE tiger.cousub;
EOF

echo "fhir_test database created"