package controller

import (
	"fmt"
	"github.com/danibix95/fdp_server/dbconn"
	"github.com/kataras/iris"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type AppController struct {
	dbc    *dbconn.DBController
	logger *log.Logger
}

func New(controlLogFile *os.File, dbLogFile *os.File) *AppController {
	appc := AppController{
		dbc: dbconn.New(dbLogFile),
		logger: log.New(io.MultiWriter(os.Stderr, controlLogFile),
			"", log.LstdFlags),
	}

	return &appc
}

func (appc *AppController) Ping(ctx iris.Context) {
	appc.dbc.PingDB() // test db connection

	_, _ = ctx.JSON(iris.Map{
		"message": fmt.Sprintf("Pong - %v", time.Now().Local()),
	})
}

/* ======= LOGIN MANAGEMENT ======= */
func (appc *AppController) RequireLogin(ctx iris.Context) {
	appc.logger.Println("User interaction with app content!")
	// TODO: implement login capabilities!
	if false {
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.EndRequest()
	} else {
		// move forward the execution of request chain
		ctx.Next()
	}
}

func (appc *AppController) Login(ctx iris.Context) {
	_, _ = ctx.JSON(iris.Map{
		"message": "Login function not implemented!",
	})
}

func (appc *AppController) Logout(ctx iris.Context) {
	_, _ = ctx.JSON(iris.Map{
		"message": "Logout function not implemented!",
	})
}

func (appc *AppController) IsAdmin(ctx iris.Context) {
	//_, _ = ctx.JSON(iris.Map{
	//	"message": "Admin check function not implemented!",
	//})
	ctx.Next()
}

/* ======= TICKETS MANAGEMENT ======= */
func (appc AppController) readPostTicket(ticket *dbconn.Ticket, ctx *iris.Context) {
	err := (*ctx).ReadJSON(ticket)
	if err != nil {
		appc.logger.Panicln(fmt.Sprintf("Error reading JSON!\n%v", err))
	}

	// check ticket is in range
	if ticket.TicketNum > dbconn.TICKETHIGH {
		appc.logger.Panicln(fmt.Sprintf("Ticket number %v is not valid!"+
			" Ticket number is not in specified range.", ticket.TicketNum))
	}
}

// Used to check whether and when an attendee entered to the party
func (appc *AppController) WhenEntered(ctx iris.Context) {
	ticketNum, err := ctx.Params().GetUint("ticketNum")
	if err != nil {
		appc.logger.Panicln(fmt.Sprintf("Ticket number %v is not valid!"+
			" Ticket numbers should be a natural number.", ticketNum))
	}
	enteredTime := appc.dbc.WhenEntered(ticketNum)

	_, _ = ctx.JSON(iris.Map{
		"ticketNum": ticketNum,
		"time":      enteredTime.Time,
		"isEntered": enteredTime.Valid,
		"status":    200,
	})
}

func (appc *AppController) GetTickets(ctx iris.Context) {
	ticketList, err := appc.dbc.TicketsList()

	result := iris.Map{
		"status":    500,
		"attendees": nil,
	}

	if err != nil {
		result["msg"] = "An error occurred while retrieving the list of attendees!"
	} else {
		result["status"] = 200
		result["attendees"] = ticketList
	}

	_, _ = ctx.JSON(result)
}

func (appc *AppController) GetTicketsStats(ctx iris.Context) {
	cEntered, cSold := make(chan int), make(chan int)

	// run the two database calls concurrently
	go appc.dbc.GetCurrentInside(cEntered)
	go appc.dbc.GetCurrentSold(cSold)

	numCurrentInside, numCurrentSold := <-cEntered, <-cSold

	result := iris.Map{
		"status":        500,
		"currentInside": numCurrentInside,
		"currentSold":   numCurrentSold,
	}

	// check that both goroutines provide a meaningful number
	if numCurrentInside > -1 && numCurrentSold > -1 {
		result["status"] = 200
	} else {
		appc.logger.Println("Error retrieving tickets stats")
	}

	_, _ = ctx.JSON(result)
}

func (appc *AppController) GetTicketDetails(ctx iris.Context) {
	ticketNum, err := ctx.Params().GetUint("ticketNum")
	if err != nil || ticketNum > dbconn.TICKETHIGH {
		appc.logger.Panicln(fmt.Sprintf("Ticket number %v is not valid!"+
			" Ticket number is not in specified range.", ticketNum))
	}

	result := iris.Map{
		"status":   400,
		"attendee": nil,
		"exists":   true,
	}

	attendee, err := appc.dbc.TicketDetails(ticketNum)
	if err != nil {
		if ticketNum > dbconn.TICKETHIGH {
			result["exists"] = false
			result["msg"] = fmt.Sprintf("Ticket do not exists."+
				"%v is outside valid range [0-%v]", ticketNum, dbconn.TICKETHIGH)
		} else {
			result["msg"] = fmt.Sprintf("Ticket %v has not been sold."+
				" Therefore no details are available", ticketNum)
		}
	} else {
		result["status"] = 200
		result["attendee"] = attendee
	}

	_, _ = ctx.JSON(result)
}

func (appc *AppController) SetEntered(ctx iris.Context) {
	var ticket dbconn.Ticket
	appc.readPostTicket(&ticket, &ctx)

	// Default result -> the ticket has not been sold
	// and therefore it cannot enter without first pay the entrance
	result := iris.Map{
		"ticketNum": ticket.TicketNum,
		"status":    400,
		"entered":   false,
		"msg":       "Ticket unsold!",
	}

	switch appc.dbc.IsSoldEntered(ticket.TicketNum) {
	case dbconn.SOLDENTERED:
		// notify that the ticket is already entered,
		// so the same ticket number cannot enter again
		result["entered"] = true
		result["msg"] = "Ticket sold and already entered."
	case dbconn.SOLD:
		if appc.dbc.SetEntered(ticket.TicketNum) {
			// successful update
			result["status"] = 200
			result["entered"] = true
			result["msg"] = "Ticket entered correctly!"

			appc.logger.Println(fmt.Sprintf("Ticket %v entered", ticket.TicketNum))
		} else {
			result["status"] = 500
			result["msg"] = "Error encountered while allowing the entrance to this ticket..."
		}
	}

	_, _ = ctx.JSON(result)
}

func (appc *AppController) SellTicket(ctx iris.Context) {
	var ticket dbconn.Ticket
	appc.readPostTicket(&ticket, &ctx)

	result := iris.Map{
		"ticketNum": ticket.TicketNum,
		"status":    500,
		"msg":       "Failed to sold selected ticket!",
	}

	// TODO: implement check to avoid overwrite ticket without explicit consent

	if ticket.FirstName == "" || ticket.LastName == "" {
		result["status"] = 400
		result["msg"] = fmt.Sprintf("Missing some atteendee details: first name -> %v, last name -> %v",
			ticket.FirstName, ticket.LastName)
	} else {
		// save in the database only capitalized names
		ticket.FirstName = strings.Title(ticket.FirstName)
		ticket.LastName = strings.Title(ticket.LastName)

		if appc.dbc.SellTicket(ticket.TicketNum, ticket.FirstName, ticket.LastName) {
			result["status"] = 200
			result["msg"] = fmt.Sprintf("Ticket sold correctly to %v %v!",
				ticket.FirstName, ticket.LastName)
		}
	}

	_, _ = ctx.JSON(result)
}

func (appc *AppController) RollbackEntrance(ctx iris.Context) {
	var ticket dbconn.Ticket
	appc.readPostTicket(&ticket, &ctx)

	result := iris.Map{
		"ticketNum": ticket.TicketNum,
		"status":    500,
		"rollback":  false,
		"msg":       "Entrance rollback failed!",
	}

	if appc.dbc.RollbackEntrance(ticket.TicketNum) {
		result["status"] = 200
		result["rollback"] = true
		result["msg"] = "Entrance rollback correctly executed!"
	}

	_, _ = ctx.JSON(result)
}

func (appc *AppController) GetTicketVendor(ctx iris.Context) {
	//ticketNum, err := ctx.Params().GetUint("ticketNum")
}

/* ======= ERROR MANAGEMENT ======= */
func (appc *AppController) Unauthorized(ctx iris.Context) {
	_, _ = ctx.JSON(iris.Map{
		"status": 401,
	})
}

func (appc *AppController) NotFound(ctx iris.Context) {
	_, _ = ctx.JSON(iris.Map{
		"status": 404,
	})
}
func (appc *AppController) InternalError(ctx iris.Context) {
	_, _ = ctx.JSON(iris.Map{
		"status": 500,
	})
}
