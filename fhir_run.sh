#!/bin/bash
cd /opt/gofhir
nohup ./gofhir -readonly -disable-ci-searches > ./gofhir.log &
