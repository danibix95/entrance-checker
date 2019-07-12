# Database Management

In this folder are contained the files to create a new Docker container which executes an instance of PostgreSQL with a specific configuration.

### Create the container

To create the container it is expected that [Docker](https://www.docker.com/) is installed in your machine and that the file `postgres-passwd` is created in this directory. This file should contain the _password_ used by the database user to access the data relevant for the application. Once it has been created, the script `create_db_image.sh` should be run to create the PostgreSQL container.

**Note**: previous command also automatically starts the container. To stop it is sufficient to run the command:

    docker stop fdp-db-docker

To subsequently launch the container, the script `launch_db.sh` is provided.

### Import data into the database

The first time that the container it is run, the application database exits, but it is empty. Therefore, by means of `import_data.sql` script is possible to create needed schema, tables and insert corresponding data.

To run the script, first change the working directory to `sql_scripts` and then launch the `psql` command (using the password contained in `postgres-passwd` file) and from it run the import script.

    psql -U postgres -h localhost -d postgres

In case it is necessary to reinitialize the database, the script `reinit_db.sql` contains the instruction to carry out the task. **Note**: run this script when connected to a database different from the one of the application (e.g. postgres).

### Manage the database

In a similar manner, to view the data contained in the application database it is necessary to run `psql` front-end with the user `fdp`.

    psql -U fdp -h localhost -d fdp_tickets