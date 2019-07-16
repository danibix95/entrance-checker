FROM postgres:11.4-alpine

WORKDIR /sql_scripts

COPY postgres-passwd /run/secrets/postgres-passwd
COPY sql_scripts/init_db.sql /sql_scripts/init_db.sql
COPY sql_scripts/init_schema.sql /sql_scripts/init_schema.sql

ENV LANG it_IT.UTF-8
ENV LANGUAGE it_IT.UTF-8
ENV LC_ALL it_IT.UTF-8

ENV POSTGRES_PASSWORD_FILE /run/secrets/postgres-passwd

# Note: name of the sql scripts are alphabetically sorted,
# so that their execution happens in the proper order
RUN echo "CREATE USER fdp WITH PASSWORD '$(cat /run/secrets/postgres-passwd)';" > /docker-entrypoint-initdb.d/fdp_user.sql
RUN cp /sql_scripts/init_db.sql /docker-entrypoint-initdb.d/init_db.sql
RUN cp /sql_scripts/init_schema.sql /docker-entrypoint-initdb.d/init_schema.sql

USER postgres