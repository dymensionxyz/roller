#!/bin/bash
random_number=$(( RANDOM % 100000 + 1 ))
# will be passed in the orchestrator script
IDENTIFIER="RAX"
/usr/local/bin/roller config init loadtest_${random_number}-1 RAX --hub froopyland --da local
echo "hello world"
