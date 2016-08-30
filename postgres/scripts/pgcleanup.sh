#!/bin/bash
# Removes the local postgres database that supports SyntheticMass statistics.
# This removes the database and roles created by pgsetup.sh.
# 
# Carlton Duffett
# 08-15-2016

psql <<EOF
SET client_encoding = 'UTF8';
DROP DATABASE IF EXISTS fhir_test;
DROP ROLE IF EXISTS fhir_test;
EOF

echo "fhir_test database removed"