\c fdp_tickets;

-- \copy rely on current directory
\copy tickets.attendees FROM '../csv/attendees.csv' WITH (FORMAT CSV, NULL 'null');
\copy tickets.fdp_staff FROM '../csv/fdp_staff.csv' WITH (FORMAT CSV);
