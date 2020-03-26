package dbconn

import "github.com/lib/pq"

const TICKETHIGH uint = 1050

type Attendee struct {
	TicketNum  uint        `json:"ticket_num"`
	LastName   string      `json:"last_name"`
	FirstName  string      `json:"firs_tname"`
	TicketType int         `json:"ticket_type"`
	Sold       bool        `json:"sold"`
	Vendor     string      `json:"vendor"`
	RespVendor string      `json:"resp_vendor"`
	Entered    pq.NullTime `json:"entered"`
}

type Staff struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Admin    bool   `json:"admin"`
}
