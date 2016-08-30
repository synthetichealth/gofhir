#!/bin/bash
# Adds the fact tables and disease table to the synth_ma schema.
# These are used to track the synthetic patient and condition
# statistics. This script is used to update the PRODUCTION DATABASE.
# Use it with caution.
# 
# Carlton Duffett
# 08-30-2016

echo "Creating synthetic facts tables..."

PWD=`pwd`

# Create the new tables and schema
psql -d fhir <<EOF
SET client_encoding = 'UTF8';

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

ALTER SEQUENCE disease_id_seq OWNER TO synth_ma;
ALTER TABLE synth_ma.synth_disease OWNER TO synth_ma;

-- Create county fact table
CREATE SEQUENCE county_factid_seq;
CREATE TABLE synth_ma.synth_county_facts (
    factid integer NOT NULL DEFAULT nextval('county_factid_seq'),
    countyfp character varying(3) NOT NULL,
    diseasefp integer NOT NULL,
    pop numeric,
    pop_male numeric,
    pop_female numeric,
    rate double precision,
    CONSTRAINT pk_county_factid PRIMARY KEY (factid)
)
WITH (
    OIDS=FALSE
);

ALTER SEQUENCE county_factid_seq OWNER TO synth_ma;
ALTER TABLE synth_ma.synth_county_facts OWNER TO synth_ma;
CREATE UNIQUE INDEX county_facts_uidx ON synth_ma.synth_county_facts (countyfp, diseasefp);

-- Create county subdivision fact table
CREATE SEQUENCE cousub_factid_seq;
CREATE TABLE synth_ma.synth_cousub_facts (
    factid integer NOT NULL DEFAULT nextval('cousub_factid_seq'),
    cousubfp character varying(5) NOT NULL,
    diseasefp integer NOT NULL,
    pop numeric,
    pop_male numeric,
    pop_female numeric,
    rate double precision,
    CONSTRAINT pk_cousub_factid PRIMARY KEY (factid)
)
WITH (
    OIDS=FALSE
);

ALTER SEQUENCE cousub_factid_seq OWNER TO synth_ma;
ALTER TABLE synth_ma.synth_cousub_facts OWNER TO synth_ma;
CREATE UNIQUE INDEX cousub_facts_uidx ON synth_ma.synth_cousub_facts (cousubfp, diseasefp);
EOF

echo "Importing current disease data..."

# Load the current disease data
cat $PWD/../data/disease.csv | psql -d fhir -c "\COPY synth_ma.synth_disease (diseasefp, stat_name, condition_name, code_icd9, code_icd10, code_snomed) FROM STDIN (DELIMITER ',', QUOTE '\"', HEADER TRUE, FORMAT CSV)"

echo "Populating fact tables..."

# Populate the fact tables with zeroed stats data.
# This is every permutation of county/cousub ID and disease ID.
# tiger_cb14_500k
psql -d fhir <<EOF
-- Populate the county fact table
INSERT INTO synth_ma.synth_county_facts(countyfp, diseasefp, pop, pop_male, pop_female, rate)
SELECT c.countyfp, d.diseasefp, 0, 0, 0, 0
FROM tiger_cb14_500k.county AS c 
CROSS JOIN synth_ma.synth_disease AS d;

-- Populate the cousub fact table
INSERT INTO synth_ma.synth_cousub_facts(cousubfp, diseasefp, pop, pop_male, pop_female, rate)
SELECT c.cousubfp, d.diseasefp, 0, 0, 0, 0
FROM tiger.cousub AS c 
CROSS JOIN synth_ma.synth_disease AS d;
EOF

echo "Cleaning up..."

# cleanup after database creation and population
psql -d fhir <<EOF
VACUUM ANALYZE synth_ma.synth_disease;
VACUUM ANALYZE synth_ma.synth_county_facts;
VACUUM ANALYZE synth_ma.synth_cousub_facts;
EOF

echo "Synthetic fact tables created."
