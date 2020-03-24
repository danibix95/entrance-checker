#! /usr/bin/bash

# create if not exist the folder containing the volume data (not needed at the moment)
# mkdir -p postgres_data

# run only the first time
docker build -t fdp_db .