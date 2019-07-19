\c fdp_tickets;

-- remove old schema to avoid using previous settings and data
DROP SCHEMA IF EXISTS tickets cascade;
CREATE SCHEMA IF NOT EXISTS tickets AUTHORIZATION fdp;
ALTER DATABASE fdp_tickets SET search_path TO tickets, public;

CREATE TABLE tickets.attendees (
    ticket_num INTEGER PRIMARY KEY,
    last_name VARCHAR,
    first_name VARCHAR,
    ticket_type SMALLINT,
    sold BOOLEAN DEFAULT FALSE,
    vendor VARCHAR,
    resp_vendor VARCHAR,
    entered timestamptz DEFAULT NULL
);

CREATE TABLE tickets.fdp_staff (
    username VARCHAR,
    password VARCHAR,
    admin BOOLEAN
);

-- assign the ownership of above table to correct user
ALTER TABLE IF EXISTS tickets.attendees OWNER TO fdp;
ALTER TABLE IF EXISTS tickets.fdp_staff OWNER TO fdp;