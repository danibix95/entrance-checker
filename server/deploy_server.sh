#! /bin/bash

docker build -t fdp_server:latest . && docker run -d -p 8080:8080 --mount source=fdp_volume,target=/app/logs --name fdp-server --rm fdp_server:latest