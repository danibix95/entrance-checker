-- initalize database envinronment
DROP DATABASE IF EXISTS fdp_tickets;
CREATE DATABASE fdp_tickets
    WITH OWNER      = fdp
         ENCODING   = 'UTF8'
         TEMPLATE   = template0
         TABLESPACE = pg_default
         LC_COLLATE = 'it_IT.UTF-8'
         LC_CTYPE   = 'it_IT.UTF-8'
         CONNECTION LIMIT = -1;

-- restore permissions for fdp user on this database
GRANT ALL PRIVILEGES ON DATABASE fdp_tickets TO fdp;