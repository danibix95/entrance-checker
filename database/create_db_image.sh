#! /usr/bin/bash

# run only the first time
echo "-- Building FdP Database --"
if [[ "$1" == "--test" ]]; then
    echo "-- DEVELOPMENT IMAGE --"
    docker build -f Dockerfile.test -t fdp_db .
else
    echo "-- PRODUCTION IMAGE --"
    docker build -t fdp_db .
fi
