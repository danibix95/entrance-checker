package controller

import (
	"fmt"
	"github.com/danibix95/fdp_server/dbconn"
	"github.com/kataras/iris"
	"io"
	"log"
	"os"
	"time"
)

// data structure for POST requests
type Ticket struct {
	TicketNum uint `json:"ticketNum"`
}

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
}

func (appc *AppController) Logout(ctx iris.Context) {
}

func (appc *AppController) IsAdmin(ctx iris.Context) {
}

/* ======= TICKETS MANAGEMENT ======= */
func (appc AppController) readPostTicket(ticket *Ticket, ctx *iris.Context) {
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
}

func (appc *AppController) GetTicketsStats(ctx iris.Context) {
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
				"Therefore no details are available", ticketNum)
		}
	} else {
		result["status"] = 200
		result["attendee"] = attendee
	}

	_, _ = ctx.JSON(result)
}

func (appc *AppController) SetEntered(ctx iris.Context) {
	var ticket Ticket
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
		} else {
			result["status"] = 500
			result["msg"] = "Error encountered while allowing the entrance to this ticket..."
		}
	}

	_, _ = ctx.JSON(result)
}

func (appc *AppController) SellTicket(ctx iris.Context) {
}

func (appc *AppController) RollbackEntrance(ctx iris.Context) {
	var ticket Ticket
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
