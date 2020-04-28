package main

import (
	"github.com/kataras/iris/httptest"
	"testing"
)

func TestPrepareApp(t *testing.T) {
	app := prepareApp(nil, nil)
	e := httptest.New(t, app)

	// redirects to /admin without basic auth
	// e.GET("/").Expect().Status(httptest.StatusUnauthorized)
	// // without basic auth
	// e.GET("/admin").Expect().Status(httptest.StatusUnauthorized)
	//
	// // with valid basic auth
	// e.GET("/admin").WithBasicAuth("myusername", "mypassword").Expect().
	// 	Status(httptest.StatusOK).Body().Equal("/admin myusername:mypassword")
	// e.GET("/admin/profile").WithBasicAuth("myusername", "mypassword").Expect().
	// 	Status(httptest.StatusOK).Body().Equal("/admin/profile myusername:mypassword")
	// e.GET("/admin/settings").WithBasicAuth("myusername", "mypassword").Expect().
	// 	Status(httptest.StatusOK).Body().Equal("/admin/settings myusername:mypassword")
	//
	// // with invalid basic auth
	// e.GET("/admin/settings").WithBasicAuth("invalidusername", "invalidpassword").
	// 	Expect().Status(httptest.StatusUnauthorized)

	e.GET("/").Expect().Status(httptest.StatusNotFound)

	// e.GET("/ping").Expect().Status(httptest.StatusUnauthorized)
	e.GET("/ping").Expect().Status(httptest.StatusOK).
		JSON().Object().Value("message").String().Raw()

	/* ========================= TEST WHEN ENTERED ========================== */
	// unauthorized access
	// e.GET("/when-entered/").Expect().Status(httptest.StatusUnauthorized)
	// e.GET("/when-entered/1").Expect().Status(httptest.StatusUnauthorized)
	// e.GET("/when-entered/42").Expect().Status(httptest.StatusUnauthorized)

	t3c1 := e.GET("/when-entered/1").Expect().Status(httptest.StatusOK).JSON().Object()
	t3c1.Value("ticketNum").Number().Equal(1)
	t3c1.Value("isEntered").Boolean().Equal(true)
	t3c1.Value("status").Number().Equal(httptest.StatusOK)

	t3c2 := e.GET("/when-entered/42").Expect().Status(httptest.StatusOK).JSON().Object()
	t3c2.Value("ticketNum").Number().Equal(42)
	t3c2.Value("time").String().Equal("0001-01-01T00:00:00Z")
	t3c2.Value("isEntered").Boolean().Equal(false)
	t3c2.Value("status").Number().Equal(httptest.StatusOK)

	// wrong ticket requests
	e.GET("/when-entered/").Expect().Status(httptest.StatusNotFound)
	e.GET("/when-entered/1920").Expect().Status(httptest.StatusNotFound)
	e.GET("/when-entered/-9").Expect().Status(httptest.StatusNotFound)
	e.GET("/when-entered/246.7").Expect().Status(httptest.StatusNotFound)
	e.GET("/when-entered/test_ticket").Expect().Status(httptest.StatusNotFound)

	/* ========================= TEST TICKETS LIST ========================== */
	// unauthorized access
	// e.GET("/tickets").Expect().Status(httptest.StatusUnauthorized)

	t4c1 := e.GET("/tickets").Expect().Status(httptest.StatusOK).JSON().Object()
	t4c1.Value("status").Number().Equal(httptest.StatusOK)
	t4c1.Value("attendees").Array().Length().Gt(0)
	t4c1.Value("attendees").Array().Length().Equal(1100)

	/* ======================== TEST TICKETS DETAILS ======================== */
	// unauthorized access
	// e.GET("/tickets/8").Expect().Status(httptest.StatusUnauthorized)

	// correct access with unsold ticket
	t5c1 := e.GET("/tickets/130").Expect().Status(httptest.StatusOK).
		JSON().Object().Value("attendee").Object()
	t5c1.Value("first_name").Object().Value("String").String().Equal("")
	t5c1.Value("last_name").Object().Value("String").String().Equal("")
	t5c1.Value("ticket_type").Number().Equal(10)
	t5c1.Value("sold").Boolean().Equal(false)
	t5c1.Value("vendor").Object().Value("Valid").Boolean().Equal(false)
	t5c1.Value("resp_vendor").Object().Value("String").String().Equal("Generator")
	t5c1.Value("entered").Object().Value("Valid").Boolean().Equal(false)

	// // correct access with ticket sold but not entered
	t5c2 := e.GET("/tickets/8").Expect().Status(httptest.StatusOK).
		JSON().Object().Value("attendee").Object()
	t5c2.Value("first_name").Object().Value("String").String().Equal("Rosa")
	t5c2.Value("last_name").Object().Value("String").String().Equal("Mora")
	t5c2.Value("ticket_type").Number().Equal(10)
	t5c2.Value("sold").Boolean().Equal(true)
	t5c2.Value("vendor").Object().Value("String").String().Equal("Ergin Schellen")
	t5c2.Value("resp_vendor").Object().Value("String").String().Equal("Generator")
	t5c2.Value("entered").Object().Value("Time").String().Equal("0001-01-01T00:00:00Z")
	t5c2.Value("entered").Object().Value("Valid").Boolean().Equal(false)

	// correct access with ticket sold and entered
	t5c3 := e.GET("/tickets/200").Expect().Status(httptest.StatusOK).
		JSON().Object().Value("attendee").Object()
	t5c3.Value("first_name").Object().Value("String").String().Equal("Isabel")
	t5c3.Value("last_name").Object().Value("String").String().Equal("Rubio")
	t5c3.Value("ticket_type").Number().Equal(10)
	t5c3.Value("sold").Boolean().Equal(true)
	t5c3.Value("vendor").Object().Value("String").String().Equal("Ergin Schellen")
	t5c3.Value("resp_vendor").Object().Value("String").String().Equal("Generator")
	// t5c3.Value("entered").Object().Value("Time").String().Equal("2020-04-28T13:18:22.306825Z")
	t5c3.Value("entered").Object().Value("Valid").Boolean().Equal(true)

	// wrong ticket number input
	e.GET("/tickets/1500").Expect().Status(httptest.StatusNotFound)
	e.GET("/tickets/-7").Expect().Status(httptest.StatusNotFound)
	e.GET("/tickets/128.3").Expect().Status(httptest.StatusNotFound)

	/* ========================= TEST TICKETS STATS ========================= */
	// unauthorized access
	// e.GET("/tickets-info").Expect().Status(httptest.StatusUnauthorized)

	// test correct ticket details retrieval
	t6c1 := e.GET("/tickets-info").Expect().Status(httptest.StatusOK).JSON().Object()
	t6c1.Value("status").Number().Equal(httptest.StatusOK)
	t6c1.Value("currentInside").Number().Equal(245)
	t6c1.Value("currentSold").Number().Equal(660)

	/* ====================== TEST SET ENTERED + ROLLBACK ====================== */
	// unauthorized access
	// e.POST("/tickets/entered").Expect().Status(httptest.StatusUnauthorized)
	// e.POST("/tickets/entered/rollback").Expect().Status(httptest.StatusUnauthorized)

	// correct entrance
	const ticket1 int = 3  // sold but not entered
	const ticket2 int = 10 // sold and already entered
	const ticket3 int = 21 // unsold
	testTicketCorrect := map[string]int{"ticket_num": ticket1}
	testTicketAgain := map[string]int{"ticket_num": ticket2}
	testTicketUnsold := map[string]int{"ticket_num": ticket3}

	t7c1 := e.POST("/tickets/entered").WithJSON(testTicketCorrect).Expect().
		Status(httptest.StatusOK).JSON().Object()
	t7c1.Value("ticketNum").Number().Equal(ticket1)
	t7c1.Value("status").Number().Equal(httptest.StatusOK)
	t7c1.Value("entered").Boolean().Equal(true)
	t7c1.Value("msg").String()

	t7c2 := e.POST("/tickets/entered").WithJSON(testTicketAgain).Expect().
		Status(httptest.StatusBadRequest).JSON().Object()
	t7c2.Value("ticketNum").Number().Equal(ticket2)
	t7c2.Value("status").Number().Equal(httptest.StatusBadRequest)
	t7c2.Value("entered").Boolean().Equal(true)
	t7c2.Value("msg").String()

	t7c3 := e.POST("/tickets/entered").WithJSON(testTicketUnsold).Expect().
		Status(httptest.StatusBadRequest).JSON().Object()
	t7c3.Value("ticketNum").Number().Equal(ticket3)
	t7c3.Value("status").Number().Equal(httptest.StatusBadRequest)
	t7c3.Value("entered").Boolean().Equal(false)
	t7c3.Value("msg").String()

	// rollback the entrance of the first ticket to restore the previous db state
	t7c4 := e.POST("/tickets/entered/rollback").WithJSON(testTicketCorrect).Expect().
		Status(httptest.StatusOK).JSON().Object()
	t7c4.Value("ticketNum").Number().Equal(ticket1)
	t7c4.Value("status").Number().Equal(httptest.StatusOK)
	t7c4.Value("rollback").Boolean().Equal(true)
	t7c4.Value("msg").String()

	// no rollback should be performed on a ticket that is not entered
	t7c5 := e.POST("/tickets/entered/rollback").WithJSON(testTicketCorrect).Expect().
		Status(httptest.StatusBadRequest).JSON().Object()
	t7c5.Value("ticketNum").Number().Equal(ticket1)
	t7c5.Value("status").Number().Equal(httptest.StatusBadRequest)
	t7c5.Value("rollback").Boolean().Equal(false)
	t7c5.Value("msg").String()

	// no rollback should occur for unsold ticket
	t7c6 := e.POST("/tickets/entered/rollback").WithJSON(testTicketUnsold).Expect().
		Status(httptest.StatusBadRequest).JSON().Object()
	t7c6.Value("ticketNum").Number().Equal(ticket3)
	t7c6.Value("status").Number().Equal(httptest.StatusBadRequest)
	t7c6.Value("rollback").Boolean().Equal(false)
	t7c6.Value("msg").String()
	//
	// e.POST("/admin/sell").Expect().Status(httptest.StatusUnauthorized)
	// e.POST("/admin/sell").Expect().Status(httptest.StatusUnauthorized)
	//
	// //e.POST("/admin/entered/undo").Expect().Status(httptest.StatusUnauthorized)
	// e.GET("/admin/ticket/vendor").Expect().Status(httptest.StatusUnauthorized)
	// e.GET("/admin/ticket/vendor").Expect().Status(httptest.StatusUnauthorized)
}
