Remember to write about the `postgres_info` file (ip and password!)

A user should create it and fill with the appropriate information

    USER=fdp
    PWD=<your-db-password>
    HOST=<ip-database-container>
    PORT=5432
    DB_NAME=fdp_tickets
    
Write how to get the ip, by means of
    
    docker inspect <container> | grep -i ipaddress

E.g.

    docker inspect fdp-db-docker | grep -i ipaddress
    
yields `172.17.0.2`