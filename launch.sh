#! /usr/bin/bash

DB_PWD=`uuidgen`
echo $DB_PWD > database/postgres-passwd

export DB_PWD=$DB_PWD

docker-compose up -d
