package dbconn

import "github.com/lib/pq"

const TICKETHIGH uint = 1050

type Attendee struct {
	TicketNum  uint        `json:"ticketnum"`
	LastName   string      `json:"lastname"`
	FirstName  string      `json:"firstname"`
	TicketType int         `json:"tickettype"`
	Sold       bool        `json:"sold"`
	Vendor     string      `json:"vendor"`
	RespVendor string      `json:"respvendor"`
	Entered    pq.NullTime `json:"entered"`
}

type Staff struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Admin    bool   `json:"admin"`
}
