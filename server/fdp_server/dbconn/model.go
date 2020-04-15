package dbconn

import (
	"database/sql"
	"github.com/lib/pq"
)

const TICKETHIGH uint = 1050

type Attendee struct {
	TicketNum  uint           `json:"ticket_num"`
	LastName   sql.NullString `json:"last_name"`
	FirstName  sql.NullString `json:"first_name"`
	TicketType int            `json:"ticket_type"`
	Sold       bool           `json:"sold"`
	Vendor     sql.NullString `json:"vendor"`
	RespVendor sql.NullString `json:"resp_vendor"`
	Entered    pq.NullTime    `json:"entered"`
}

type AttendeeSimple struct {
	TicketNum  uint        `json:"ticket_num"`
	TicketType int         `json:"ticket_type"`
	Sold       bool        `json:"sold"`
	Entered    pq.NullTime `json:"entered"`
}

type Staff struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Admin    bool   `json:"admin"`
}
