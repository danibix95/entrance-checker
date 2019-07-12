#! /usr/bin/bash

# create if not exist the folder containing the volume
mkdir -p ./docker/volumes/postgres

# run only the first time
docker build -t fdp_db .

# start the container with specific parameters
./launch_db.sh