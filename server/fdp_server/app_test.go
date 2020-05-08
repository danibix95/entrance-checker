package main

import (
	"fmt"
	"github.com/danibix95/fdp_server/controller"
	"github.com/danibix95/fdp_server/dbconn"
	"github.com/kataras/iris/v12/httptest"
	"testing"
)

func TestPrepareApp(t *testing.T) {
	testConf := controller.AppConfig{SecretsFile: ".app_secrets"}

	app := prepareApp(testConf)
	e := httptest.New(t, app)

	e.GET("/").Expect().Status(httptest.StatusNotFound)

	/* ============================= TEST AUTH ============================== */
	wrongLogin1 := dbconn.Login{
		Username: "adm",
		Password: "",
	}
	wrongLogin2 := dbconn.Login{
		Username: "hello",
		Password: "fkjrbgbrg",
	}
	wrongLogin3 := dbconn.Login{
		Username: "",
		Password: "fkjrbgbrg",
	}
	userLogin := dbconn.Login{
		Username: "fdp17",
		Password: "1ZLVXk6/27zEBgTGPZ3sKED3kFC5t+itqZ5SIC2QoEw=",
	}
	adminLogin := dbconn.Login{
		Username: "adm",
		Password: "POvIqSFGMWbxCdwx69HtJ+xOiav3GvnjylfGBmrwokM=",
	}

	// incorrect login credentials
	e.POST("/login").Expect().Status(httptest.StatusInternalServerError)
	e.POST("/login").WithJSON("").Expect().Status(httptest.StatusInternalServerError)
	e.POST("/login").WithJSON(wrongLogin1).Expect().Status(httptest.StatusUnauthorized)
	e.POST("/login").WithJSON(wrongLogin2).Expect().Status(httptest.StatusUnauthorized)
	e.POST("/login").WithJSON(wrongLogin3).Expect().Status(httptest.StatusUnauthorized)

	// correct entrance
	userRes := e.POST("/login").WithJSON(userLogin).Expect().
		Status(httptest.StatusOK).JSON().Object()
	userRes.Value("status").Number().Equal(httptest.StatusOK)
	userRes.Value("msg").String()
	userRes.Value("token").String()
	userToken := fmt.Sprintf("Bearer %v", userRes.Value("token").Raw())

	adminRes := e.POST("/login").WithJSON(adminLogin).Expect().
		Status(httptest.StatusOK).JSON().Object()
	adminRes.Value("status").Number().Equal(httptest.StatusOK)
	adminRes.Value("msg").String()
	adminRes.Value("token").String()
	adminToken := fmt.Sprintf("Bearer %v", adminRes.Value("token").Raw())

	/* ============================ TEST DB PING ============================ */
	e.GET("/ping").Expect().Status(httptest.StatusUnauthorized)

	e.GET("/ping").WithHeader("Authorization", userToken).Expect().
		Status(httptest.StatusOK).JSON().Object().Value("message").String()

	/* ========================= TEST WHEN ENTERED ========================== */
	// unauthorized access
	e.GET("/when-entered/2").Expect().Status(httptest.StatusUnauthorized)
	e.GET("/when-entered/42").Expect().Status(httptest.StatusUnauthorized)

	// ticket entered
	t3c1 := e.GET("/when-entered/2").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusOK).JSON().Object()
	t3c1.Value("ticketNum").Number().Equal(2)
	t3c1.Value("isEntered").Boolean().True()
	t3c1.Value("status").Number().Equal(httptest.StatusOK)

	// ticket not entered
	t3c2 := e.GET("/when-entered/42").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusOK).JSON().Object()
	t3c2.Value("ticketNum").Number().Equal(42)
	t3c2.Value("time").String().Equal("0001-01-01T00:00:00Z")
	t3c2.Value("isEntered").Boolean().False()
	t3c2.Value("status").Number().Equal(httptest.StatusOK)

	// wrong ticket requests
	e.GET("/when-entered").Expect().Status(httptest.StatusNotFound)
	e.GET("/when-entered/").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusNotFound)
	e.GET("/when-entered/1920").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusNotFound)
	e.GET("/when-entered/-9").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusNotFound)
	e.GET("/when-entered/246.7").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusNotFound)
	e.GET("/when-entered/test_ticket").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusNotFound)

	/* ========================= TEST TICKETS LIST ========================== */
	// unauthorized access
	e.GET("/tickets").Expect().Status(httptest.StatusUnauthorized)

	t4c1 := e.GET("/tickets").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusOK).JSON().Object()
	t4c1.Value("status").Number().Equal(httptest.StatusOK)
	t4c1.Value("attendees").Array().Length().Gt(0)
	t4c1.Value("attendees").Array().Length().Equal(1050)

	/* ======================== TEST TICKETS DETAILS ======================== */
	// unauthorized access
	e.GET("/tickets/8").Expect().Status(httptest.StatusUnauthorized)

	// unsold ticket
	t5c1 := e.GET("/tickets/230").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusOK).JSON().Object().Value("attendee").Object()
	t5c1.Value("first_name").Object().Value("String").String().Equal("")
	t5c1.Value("last_name").Object().Value("String").String().Equal("")
	t5c1.Value("ticket_type").Number().Equal(10)
	t5c1.Value("sold").Boolean().False()
	t5c1.Value("vendor").Object().Value("Valid").Boolean().False()
	t5c1.Value("resp_vendor").Object().Value("String").String().Equal("Generator")
	t5c1.Value("entered").Object().Value("Valid").Boolean().False()

	// ticket sold but not entered
	t5c2 := e.GET("/tickets/125").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusOK).JSON().Object().Value("attendee").Object()
	t5c2.Value("first_name").Object().Value("String").String().Equal("Silvia")
	t5c2.Value("last_name").Object().Value("String").String().Equal("Fuentes")
	t5c2.Value("ticket_type").Number().Equal(10)
	t5c2.Value("sold").Boolean().True()
	t5c2.Value("vendor").Object().Value("String").String().Equal("Ergin Schellen")
	t5c2.Value("resp_vendor").Object().Value("String").String().Equal("Generator")
	t5c2.Value("entered").Object().Value("Time").String().Equal("0001-01-01T00:00:00Z")
	t5c2.Value("entered").Object().Value("Valid").Boolean().False()

	// ticket sold and entered
	t5c3 := e.GET("/tickets/4").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusOK).JSON().Object().Value("attendee").Object()
	t5c3.Value("first_name").Object().Value("String").String().Equal("Florentine")
	t5c3.Value("last_name").Object().Value("String").String().Equal("Kost")
	t5c3.Value("ticket_type").Number().Equal(10)
	t5c3.Value("sold").Boolean().True()
	t5c3.Value("vendor").Object().Value("String").String().Equal("Zenab Creemers")
	t5c3.Value("resp_vendor").Object().Value("String").String().Equal("Generator")
	t5c3.Value("entered").Object().Value("Valid").Boolean().True()

	// corner case: ticket sold but not paid
	t5c4 := e.GET("/tickets/82").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusOK).JSON().Object().Value("attendee").Object()
	t5c4.Value("first_name").Object().Value("String").String().Equal("Naömi")
	t5c4.Value("last_name").Object().Value("String").String().Equal("Van gulik")
	t5c4.Value("ticket_type").Number().Equal(10)
	t5c4.Value("sold").Boolean().False()
	t5c4.Value("vendor").Object().Value("String").String().Equal("Ergin Schellen")
	t5c4.Value("resp_vendor").Object().Value("String").String().Equal("Generator")
	t5c4.Value("entered").Object().Value("Time").String().Equal("0001-01-01T00:00:00Z")
	t5c4.Value("entered").Object().Value("Valid").Boolean().False()

	// wrong ticket number input
	e.GET("/tickets/1500").Expect().Status(httptest.StatusNotFound)
	e.GET("/tickets/1500").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusNotFound)
	e.GET("/tickets/-7").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusNotFound)
	e.GET("/tickets/128.3").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusNotFound)

	/* ========================= TEST TICKETS STATS ========================= */
	// unauthorized access
	e.GET("/tickets-info").Expect().Status(httptest.StatusUnauthorized)

	// test correct ticket details retrieval
	t6c1 := e.GET("/tickets-info").WithHeader("Authorization", userToken).
		Expect().Status(httptest.StatusOK).JSON().Object()
	t6c1.Value("status").Number().Equal(httptest.StatusOK)
	t6c1.Value("currentInside").Number().Equal(226)
	t6c1.Value("currentSold").Number().Equal(688)
	t6c1.Value("currentPayingEntered").Number().Equal(185)

	/* ==================== TEST SET ENTERED + ROLLBACK ===================== */
	// unauthorized access
	e.POST("/tickets/entered").Expect().Status(httptest.StatusUnauthorized)
	e.POST("/tickets/entered/rollback").Expect().Status(httptest.StatusUnauthorized)

	tTicketSold := dbconn.Ticket{TicketNum: 125}  // sold but not entered
	tTicketEntered := dbconn.Ticket{TicketNum: 4} // sold and already entered
	tTicketUnsold := dbconn.Ticket{TicketNum: 30} // unsold
	tTicketUnpaid := dbconn.Ticket{TicketNum: 82} // unpaid

	// correct entrance
	t7c1 := e.POST("/tickets/entered").WithHeader("Authorization", userToken).
		WithJSON(tTicketSold).Expect().Status(httptest.StatusOK).JSON().Object()
	t7c1.Value("ticketNum").Number().Equal(tTicketSold.TicketNum)
	t7c1.Value("status").Number().Equal(httptest.StatusOK)
	t7c1.Value("entered").Boolean().True()
	t7c1.Value("msg").String()

	// try to set entered a ticket that is already entered
	t7c2 := e.POST("/tickets/entered").WithHeader("Authorization", userToken).
		WithJSON(tTicketEntered).Expect().Status(httptest.StatusBadRequest).JSON().Object()
	t7c2.Value("ticketNum").Number().Equal(tTicketEntered.TicketNum)
	t7c2.Value("status").Number().Equal(httptest.StatusBadRequest)
	t7c2.Value("entered").Boolean().True()
	t7c2.Value("msg").String()

	// try to set entered a ticket that has not been sold
	t7c3 := e.POST("/tickets/entered").WithHeader("Authorization", userToken).
		WithJSON(tTicketUnsold).Expect().Status(httptest.StatusBadRequest).JSON().Object()
	t7c3.Value("ticketNum").Number().Equal(tTicketUnsold.TicketNum)
	t7c3.Value("status").Number().Equal(httptest.StatusBadRequest)
	t7c3.Value("entered").Boolean().False()
	t7c3.Value("msg").String()

	// try to set entered a ticket that has not been paid
	t7c4 := e.POST("/tickets/entered").WithHeader("Authorization", userToken).
		WithJSON(tTicketUnpaid).Expect().Status(httptest.StatusBadRequest).JSON().Object()
	t7c4.Value("ticketNum").Number().Equal(tTicketUnpaid.TicketNum)
	t7c4.Value("status").Number().Equal(httptest.StatusBadRequest)
	t7c4.Value("entered").Boolean().False()
	t7c4.Value("msg").String()

	// rollback the entrance of the first ticket to restore the previous db state
	t7c5 := e.POST("/tickets/entered/rollback").WithHeader("Authorization", userToken).
		WithJSON(tTicketSold).Expect().Status(httptest.StatusOK).JSON().Object()
	t7c5.Value("ticketNum").Number().Equal(tTicketSold.TicketNum)
	t7c5.Value("status").Number().Equal(httptest.StatusOK)
	t7c5.Value("rollback").Boolean().True()
	t7c5.Value("msg").String()

	// no rollback should be performed on a ticket that is not entered
	t7c6 := e.POST("/tickets/entered/rollback").WithHeader("Authorization", userToken).
		WithJSON(tTicketSold).Expect().Status(httptest.StatusBadRequest).JSON().Object()
	t7c6.Value("ticketNum").Number().Equal(tTicketSold.TicketNum)
	t7c6.Value("status").Number().Equal(httptest.StatusBadRequest)
	t7c6.Value("rollback").Boolean().False()
	t7c6.Value("msg").String()

	// no rollback should occur for unsold ticket
	t7c7 := e.POST("/tickets/entered/rollback").WithHeader("Authorization", userToken).
		WithJSON(tTicketUnsold).Expect().Status(httptest.StatusBadRequest).JSON().Object()
	t7c7.Value("ticketNum").Number().Equal(tTicketUnsold.TicketNum)
	t7c7.Value("status").Number().Equal(httptest.StatusBadRequest)
	t7c7.Value("rollback").Boolean().False()
	t7c7.Value("msg").String()

	/* ===================== TEST SELL TICKET AND RESET ===================== */
	e.POST("/admin/sell").Expect().Status(httptest.StatusUnauthorized)
	e.POST("/admin/reset").Expect().Status(httptest.StatusUnauthorized)

	tTicketErr := dbconn.Ticket{TicketNum: 30, FirstName: "Han"}

	// update ticket details to execute the tests
	tTicketUnsold.FirstName = "Han"
	tTicketUnsold.LastName = "solo"

	tTicketSold.FirstName = "Han"
	tTicketSold.LastName = "Solo"

	tTicketEntered.FirstName = "Han"
	tTicketEntered.LastName = "Solo"

	tTicketUnpaid.FirstName = "Han"
	tTicketUnpaid.LastName = "Solo"

	// impossible to sell a ticket with missing details
	t8c1 := e.POST("/admin/sell").WithHeader("Authorization", adminToken).
		WithJSON(tTicketErr).Expect().Status(httptest.StatusBadRequest).JSON().Object()
	t8c1.Value("ticketNum").Number().Equal(tTicketErr.TicketNum)
	t8c1.Value("status").Number().Equal(httptest.StatusBadRequest)
	t8c1.Value("soldNow").Boolean().False()
	t8c1.Value("entered").Boolean().False()
	t8c1.Value("msg").String()

	// correctly sell an empty ticket
	t8c2 := e.POST("/admin/sell").WithHeader("Authorization", adminToken).
		WithJSON(tTicketUnsold).Expect().Status(httptest.StatusOK).JSON().Object()
	t8c2.Value("ticketNum").Number().Equal(tTicketUnsold.TicketNum)
	t8c2.Value("status").Number().Equal(httptest.StatusOK)
	t8c2.Value("soldNow").Boolean().True()
	t8c2.Value("entered").Boolean().True()
	t8c2.Value("msg").String().Equal("ticket sold correctly to Han Solo")

	// try to sell a ticket already sold
	t8c3 := e.POST("/admin/sell").WithHeader("Authorization", adminToken).
		WithJSON(tTicketSold).Expect().Status(httptest.StatusBadRequest).JSON().Object()
	t8c3.Value("ticketNum").Number().Equal(tTicketSold.TicketNum)
	t8c3.Value("status").Number().Equal(httptest.StatusBadRequest)
	t8c3.Value("soldNow").Boolean().False()
	t8c3.Value("entered").Boolean().False()
	t8c3.Value("msg").String()

	// try to sell a ticket already sold and entered
	t8c4 := e.POST("/admin/sell").WithHeader("Authorization", adminToken).
		WithJSON(tTicketEntered).Expect().Status(httptest.StatusBadRequest).JSON().Object()
	t8c4.Value("ticketNum").Number().Equal(tTicketEntered.TicketNum)
	t8c4.Value("status").Number().Equal(httptest.StatusBadRequest)
	t8c4.Value("soldNow").Boolean().False()
	t8c4.Value("entered").Boolean().True()
	t8c4.Value("msg").String()

	// try to sell an unpaid ticket -> ticket reserved / not available
	t8c5 := e.POST("/admin/sell").WithHeader("Authorization", adminToken).
		WithJSON(tTicketUnpaid).Expect().Status(httptest.StatusBadRequest).JSON().Object()
	t8c5.Value("ticketNum").Number().Equal(tTicketUnpaid.TicketNum)
	t8c5.Value("status").Number().Equal(httptest.StatusBadRequest)
	t8c5.Value("soldNow").Boolean().False()
	t8c5.Value("entered").Boolean().False()
	t8c5.Value("msg").String()

	// reset the ticket sold in the previous test to get back original system state
	t8c6 := e.POST("/admin/reset").WithHeader("Authorization", adminToken).
		WithJSON(tTicketUnsold).Expect().Status(httptest.StatusOK).JSON().Object()
	t8c6.Value("ticketNum").Number().Equal(tTicketUnsold.TicketNum)
	t8c6.Value("status").Number().Equal(httptest.StatusOK)
	t8c6.Value("msg").String()

	/* ======================= TEST GET TICKET VENDOR ======================= */
	e.GET("/admin/vendor/").Expect().Status(httptest.StatusNotFound)
	e.GET("/admin/vendor/82").Expect().Status(httptest.StatusUnauthorized)

	t9c1 := e.GET("/admin/vendor/82").WithHeader("Authorization", adminToken).
		Expect().Status(httptest.StatusOK).JSON().Object()
	t9c1.Value("status").Number().Equal(httptest.StatusOK)
	t9c1.Value("vendor").Object().Value("Valid").Boolean().Equal(true)
	t9c1.Value("vendor").Object().Value("String").String().Equal("Ergin Schellen")

	// wrong ticket requests
	e.GET("/admin/vendor/-5").WithHeader("Authorization", adminToken).
		Expect().Status(httptest.StatusNotFound)
	e.GET("/admin/vendor/1500").WithHeader("Authorization", adminToken).
		Expect().Status(httptest.StatusNotFound)
	e.GET("/admin/vendor/451.0").WithHeader("Authorization", adminToken).
		Expect().Status(httptest.StatusNotFound)
}
