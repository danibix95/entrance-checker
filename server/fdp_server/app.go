package main

import (
	"github.com/danibix95/fdp_server/controller"
	"github.com/danibix95/fdp_server/dbconn"
	"github.com/kataras/iris"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

const logDir string = "logs"

func prepareApp(contLog, dbLog *os.File) *iris.Application {
	// Before starting the application,
	// obtain a controller for it, which connects to the database
	control := controller.New(contLog, dbLog)

	// APP DEFINITION
	app := iris.Default()

	app.Post("/login", control.Login)

	privateRoutes := app.Party("/", control.RequireLogin)
	{
		privateRoutes.Get("/ping", control.Ping)
		privateRoutes.Get("/when-entered/{ticketNum:uint max("+
			strconv.FormatUint(uint64(dbconn.TICKETHIGH), 10)+")}",
			control.WhenEntered)

		// list tickets status
		privateRoutes.Get("/tickets", control.GetTickets)
		// get a notification with tickets info (entered vs sold)
		privateRoutes.Get("/tickets-info", control.GetTicketsStats)
		// get specific ticket status
		privateRoutes.Get("/tickets/{ticketNum:uint max("+
			strconv.FormatUint(uint64(dbconn.TICKETHIGH), 10)+")}",
			control.GetTicketDetails)

		// set an user as entered to the event
		privateRoutes.Post("/tickets/entered", control.SetEntered)
		// used to confirm the entrance of attendee
		privateRoutes.Post("/tickets/entered/rollback", control.RollbackEntrance)

		/* ======= ADMIN AREA =======*/
		adminRoutes := privateRoutes.Party("/admin", control.IsAdmin)
		{
			// add a ticket to dataset
			adminRoutes.Post("/sell", control.SellTicket)
			// UI way to remove an entrance
			adminRoutes.Post("/entered/undo", control.RollbackEntrance)
			// Get who sold specified ticket
			adminRoutes.Get("/ticket/vendor", control.GetTicketVendor)
		}
	}

	// Register custom handler for specific http errors.
	app.OnErrorCode(iris.StatusUnauthorized, control.Unauthorized)
	app.OnErrorCode(iris.StatusNotFound, control.NotFound)
	app.OnErrorCode(iris.StatusInternalServerError, control.InternalError)

	return app
}

func main() {
	// to avoid potential errors, create the logs directory if it does not exists
	if _, err := os.Stat("logs"); err != nil {
		if os.IsNotExist(err) {
			dirErr := os.Mkdir("logs", os.ModeDir)
			if dirErr != nil {
				panic("Impossible to find or create logs folder!")
			}
		}
	}

	// log files
	controlLogFile, err := os.OpenFile(filepath.Join(logDir, "control.log"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer controlLogFile.Close()

	dbLogFile, err := os.OpenFile(filepath.Join(logDir, "db_conn.log"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer dbLogFile.Close()

	app := prepareApp(controlLogFile, dbLogFile)

	// with _ = it is possible to ignore the value returned by the method
	_ = app.Run(iris.Addr(":8080"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithCharset("UTF-8"))
}
