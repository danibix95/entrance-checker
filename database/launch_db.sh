#! /usr/bin/bash
docker run -d -p 5432:5432 -v "$(pwd)/docker/volumes/postgres" --name fdp-db-docker fdp_db:latest