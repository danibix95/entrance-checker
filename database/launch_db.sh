#! /usr/bin/bash

# Note: this lauch configuration does not set a specific place for saving
# the database data, but it rather let Docker create the volume in an appropriate location
#
# For more control is possible to select the volume with the following flag:
#
#        -v <your-selected-location>:/var/lib/postgresql/data
#
# but it currently does not work due to issue with file permissions / SELinux

docker run --rm -d -p 5432:5432 -h db --name fdp-db-docker fdp_db:latest
