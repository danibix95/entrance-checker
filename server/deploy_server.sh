#! /bin/bash

docker build -t fdp_server:latest .
docker run -p 8080:8080 -d -v $PWD/logs,target=/app/logs --name fdp-server --rm fdp_server:latest