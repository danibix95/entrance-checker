FROM postgres:15.2-alpine as database_image

WORKDIR /sql_scripts

COPY postgres-passwd /run/secrets/postgres-passwd
COPY sql_scripts/ .

RUN mkdir /csv
# Note: staff only for development
COPY csv/fdp_staff.csv /csv
COPY csv/attendees.csv /csv

ENV LANG it_IT.UTF-8
ENV LANGUAGE it_IT.UTF-8
ENV LC_ALL it_IT.UTF-8

ENV POSTGRES_PASSWORD_FILE /run/secrets/postgres-passwd

# Note: name of the sql scripts are alphabetically sorted,
# so that their execution happens in the proper order
RUN echo "CREATE USER fdp WITH PASSWORD '$(cat /run/secrets/postgres-passwd)';" > /docker-entrypoint-initdb.d/fdp_user.sql
RUN cp /sql_scripts/init_db.sql /docker-entrypoint-initdb.d/init_db.sql
RUN cp /sql_scripts/init_schema.sql /docker-entrypoint-initdb.d/init_schema.sql
RUN cp /sql_scripts/import_data.sql /docker-entrypoint-initdb.d/insert_data.sql

USER postgres
