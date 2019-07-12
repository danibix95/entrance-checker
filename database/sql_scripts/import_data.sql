\c fdp_tickets;

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
    canSell BOOLEAN
);

-- \copy rely on current directory
\copy tickets.attendees FROM 'attendees.csv' WITH (FORMAT CSV, NULL 'null');
\copy tickets.fdp_staff FROM 'fdp_staff.csv' WITH (FORMAT CSV);

\c fdp_tickets;