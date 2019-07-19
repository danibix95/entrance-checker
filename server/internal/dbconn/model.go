package dbconn

import "github.com/lib/pq"

type Attendee struct {
	TicketNum  uint
	LastName   string
	FirstName  string
	TicketType int
	Sold       bool
	Vendor     string
	RespVendor string
	Entered    pq.NullTime
}

type Staff struct {
	Username string
	Password string
	Admin    bool
}
