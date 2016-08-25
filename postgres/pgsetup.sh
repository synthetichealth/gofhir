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
-- Add the PostGIS database extensions
CREATE EXTENSION postgis;
CREATE EXTENSION fuzzystrmatch;
CREATE EXTENSION postgis_tiger_geocoder;

-- Create roles
CREATE USER fhir_test WITH PASSWORD 'fhir_test';

-- Create schema
CREATE SCHEMA synth_ma AUTHORIZATION fhir_test;
-- NOTE: The "tiger" schema is automatically created by PostGIS.
GRANT USAGE ON SCHEMA tiger TO public;
GRANT SELECT ON TABLE tiger.county TO public;
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
)
WITH (
    OIDS=FALSE
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
)
WITH (
    OIDS=FALSE
);
ALTER TABLE synth_ma.synth_cousub_stats OWNER TO fhir_test;

-- Create disease table
CREATE SEQUENCE disease_id_seq;
CREATE TABLE synth_ma.synth_disease (
    diseasefp integer NOT NULL DEFAULT nextval('disease_id_seq'),
    stat_name character varying(100),
    condition_name character varying(100) UNIQUE,
    code_icd9 character varying(6),
    code_icd10 character varying(6),
    code_snomed character varying(8),
    CONSTRAINT pk_diseasefp PRIMARY KEY (diseasefp)
)
WITH (
    OIDS=FALSE
);

ALTER SEQUENCE disease_id_seq OWNER TO fhir_test;
ALTER TABLE synth_ma.synth_disease OWNER TO fhir_test;

-- Create county fact table
CREATE SEQUENCE county_factid_seq;
CREATE TABLE synth_ma.synth_county_facts (
    factid integer NOT NULL DEFAULT nextval('county_factid_seq'),
    cntyidfp character varying(5) NOT NULL,
    diseasefp integer NOT NULL,
    pop numeric,
    pop_male numeric,
    pop_female numeric,
    CONSTRAINT pk_county_factid PRIMARY KEY (factid)
)
WITH (
    OIDS=FALSE
);

ALTER SEQUENCE county_factid_seq OWNER TO fhir_test;
ALTER TABLE synth_ma.synth_county_facts OWNER TO fhir_test;
CREATE UNIQUE INDEX county_facts_uidx ON synth_ma.synth_county_facts (cntyidfp, diseasefp);

-- Create county subdivision fact table
CREATE SEQUENCE cousub_factid_seq;
CREATE TABLE synth_ma.synth_cousub_facts (
    factid integer NOT NULL DEFAULT nextval('cousub_factid_seq'),
    cosbidfp character varying(10) NOT NULL,
    diseasefp integer NOT NULL,
    pop numeric,
    pop_male numeric,
    pop_female numeric,
    CONSTRAINT pk_cousub_factid PRIMARY KEY (factid)
)
WITH (
    OIDS=FALSE
);

ALTER SEQUENCE cousub_factid_seq OWNER TO fhir_test;
ALTER TABLE synth_ma.synth_cousub_facts OWNER TO fhir_test;
CREATE UNIQUE INDEX cousub_facts_uidx ON synth_ma.synth_cousub_facts (cosbidfp, diseasefp);
EOF

# County Tiger Data
# From: tiger_cb14_500k.county
# Query:
# SELECT gid, statefp, countyfp, countyns, statefp || countyfp as cntyidfp, name, lsad, aland, awater, the_geom
# FROM tiger_cb14_500k.county WHERE statefp = '25';
cat $PWD/data/county.csv | psql -d fhir_test -c "\COPY tiger.county (gid, statefp, countyfp, countyns, cntyidfp, name, lsad, aland, awater, the_geom) FROM STDIN (DELIMITER ',', QUOTE '\"', HEADER TRUE, FORMAT CSV)"

# Cousub Tiger Data
# From: tiger_cb14_500k.cousub
# Query:
# SELECT gid, statefp, countyfp, cousubfp, cousubns, statefp || countyfp || cousubfp as cosbidfp, name, lsad, aland, awater, the_geom
#  FROM tiger_cb14_500k.cousub WHERE statefp = '25';
cat $PWD/data/cousub.csv | psql -d fhir_test -c "\COPY tiger.cousub (gid, statefp, countyfp, cousubfp, cousubns, cosbidfp, name, lsad, aland, awater, the_geom) FROM STDIN (DELIMITER ',', QUOTE '\"', HEADER TRUE, FORMAT CSV)"

# Synthetic County Statistics
cat $PWD/data/synth_county_stats.csv | psql -d fhir_test -c "\COPY synth_ma.synth_county_stats (ct_name, ct_fips, sq_mi, pop, pop_male, pop_female, pop_sm, ct_poly, ct_pnt) FROM STDIN (DELIMITER ',', QUOTE '\"', HEADER TRUE, FORMAT CSV)"

# Synthetic Subdivision Statistics
cat $PWD/data/synth_cousub_stats.csv | psql -d fhir_test -c "\COPY synth_ma.synth_cousub_stats (ct_name, ct_fips, cs_name, cs_fips, sq_mi, pop, pop_male, pop_female, pop_sm, cs_poly, cs_pnt) FROM STDIN (DELIMITER ',', QUOTE '\"', HEADER TRUE, FORMAT CSV)"

# Disease Data
cat $PWD/data/disease.csv | psql -d fhir_test -c "\COPY synth_ma.synth_disease (diseasefp, stat_name, condition_name, code_icd9, code_icd10, code_snomed) FROM STDIN (DELIMITER ',', QUOTE '\"', HEADER TRUE, FORMAT CSV)"

# Populate the fact tables with zeroed stats data.
# This is every permutation of county/cousub ID and disease ID.
psql -d fhir_test <<EOF
-- Populate the county fact table
INSERT INTO synth_ma.synth_county_facts(cntyidfp, diseasefp, pop, pop_male, pop_female)
SELECT c.cntyidfp, d.diseasefp, 0, 0, 0
FROM tiger.county AS c 
CROSS JOIN synth_ma.synth_disease AS d;

-- Populate the cousub fact table
INSERT INTO synth_ma.synth_cousub_facts(cosbidfp, diseasefp, pop, pop_male, pop_female)
SELECT c.cosbidfp, d.diseasefp, 0, 0, 0
FROM tiger.cousub AS c 
CROSS JOIN synth_ma.synth_disease AS d;
EOF

# cleanup after database creation and population
psql -d fhir_test <<EOF
VACUUM ANALYZE tiger.county;
VACUUM ANALYZE tiger.cousub;
VACUUM ANALYZE synth_ma.synth_county_stats;
VACUUM ANALYZE synth_ma.synth_cousub_stats;
VACUUM ANALYZE synth_ma.synth_disease;
VACUUM ANALYZE synth_ma.synth_county_facts;
VACUUM ANALYZE synth_ma.synth_cousub_facts;
EOF

echo "fhir_test database created"