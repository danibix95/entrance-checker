FROM postgres:11.4-alpine

WORKDIR /sql_scripts

COPY postgres-passwd /run/secrets/postgres-passwd
COPY sql_scripts /sql_scripts

ENV LANG it_IT.UTF-8
ENV LANGUAGE it_IT.UTF-8
ENV LC_ALL it_IT.UTF-8

ENV POSTGRES_PASSWORD_FILE /run/secrets/postgres-passwd

RUN echo CREATE USER fdp WITH PASSWORD "'$(cat /run/secrets/postgres-passwd)';" > /docker-entrypoint-initdb.d/fdp_user.sql
RUN cp /sql_scripts/reinit_db.sql /docker-entrypoint-initdb.d/reinit_db.sql

USER postgres