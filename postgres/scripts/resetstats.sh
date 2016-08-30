#!/bin/bash
# Resets the synth_ma.synth_* stats and facts tables. This zeroes
# out any existing demographic stats in existing rows.
# This script is used to update the PRODUCTION DATABASE.
# Use it with caution.
# 
# Carlton Duffett
# 08-30-2016

echo "Resetting synthetic statistics..."

psql -d fhir <<EOF

-- Reset synth_ma.synth_county_stats
UPDATE synth_ma.synth_county_stats SET pop=0, pop_male=0, pop_female=0, pop_sm=0;

-- Reset synth_ma.synth_cousub_stats
UPDATE synth_ma.synth_cousub_stats SET pop=0, pop_male=0, pop_female=0, pop_sm=0;

-- Reset synth_ma.synth_county_facts
UPDATE synth_ma.synth_county_facts SET pop=0, pop_male=0, pop_female=0;

-- Reset synth_ma.synth_cousub_facts
UPDATE synth_ma.synth_cousub_facts SET pop=0, pop_male=0, pop_female=0;
EOF

echo "Synthetic statistics reset."