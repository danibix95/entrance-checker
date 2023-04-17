# FdP_tickets
Tickets manager for Festa di Primavera event, held at Istituto Salesiano Maria Ausiliatrice, Trento


TODO: improve instructions

### Set up

In `database` folder create new folder `csv` and insert the following files:

- `attendees.csv`
- `fdp_staff.csv`

Then execute `launch.sh` script that can be found on the root folder. This script will use `docker-compose` to build and start the containers necessary to run the applications. Once started it is possible to access the application at [localhost:52017](http://localhost:52017)