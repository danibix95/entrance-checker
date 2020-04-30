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
		logger: log.New(io.MultiWriter(controlLogFile),
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
	// TODO: implement login capabilities!
	if false {
		appc.Unauthorized(ctx)
		ctx.EndRequest()
	} else {
		// move forward the execution of request chain
		ctx.Next()
	}
}

func (appc *AppController) Login(ctx iris.Context) {
	appc.UnauthorizedMessage(ctx, "login function not implemented yet")
}

func (appc *AppController) Logout(ctx iris.Context) {
	appc.BadRequestMessage(ctx, "logout function not implemented yet")
}

func (appc *AppController) IsAdmin(ctx iris.Context) {
	ctx.Next() // ignore for the moment this function
}

/* ======= TICKETS MANAGEMENT ======= */
func (appc *AppController) readPostTicket(ctx iris.Context, ticket *dbconn.Ticket) error {
	err := ctx.ReadJSON(ticket)
	if err != nil {
		msg := fmt.Sprintf("impossible to parse request's JSON data as ticket schema")
		appc.InternalErrorMessage(ctx, msg)
	}

	// check ticket is in range
	if ticket.TicketNum > dbconn.TICKETHIGH {
		msg := fmt.Sprintf("ticket number %v is not in the correct range [%v, %v]",
			ticket.TicketNum, dbconn.TICKETLOW, dbconn.TICKETHIGH)
		appc.BadRequestMessage(ctx, msg)
	}

	return err
}

// Used to check whether and when an attendee entered to the party
func (appc *AppController) WhenEntered(ctx iris.Context) {
	ticketNum, err := ctx.Params().GetUint("ticketNum")
	if err != nil {
		msg := fmt.Sprintf("ticket number %v is not valid", ticketNum)
		appc.BadRequestMessage(ctx, msg)
	}
	enteredTime, err := appc.dbc.WhenEntered(ticketNum)

	if err != nil {
		appc.InternalErrorMessage(ctx, err.Error())
	} else {
		_, _ = ctx.JSON(iris.Map{
			"ticketNum": ticketNum,
			"time":      enteredTime.Time,
			"isEntered": enteredTime.Valid,
			"status":    iris.StatusOK,
		})
	}
}

func (appc *AppController) GetTickets(ctx iris.Context) {
	ticketList, err := appc.dbc.TicketsList()

	if err != nil {
		appc.InternalErrorMessage(ctx, err.Error())
	} else {
		_, _ = ctx.JSON(iris.Map{
			"status":    iris.StatusOK,
			"attendees": ticketList,
		})
	}
}

func (appc *AppController) GetTicketsStats(ctx iris.Context) {
	cEntered, cSold, cEnteredPaying := make(chan int), make(chan int), make(chan int)

	// run the two database calls concurrently
	go appc.dbc.GetCurrentInside(cEntered)
	go appc.dbc.GetCurrentSold(cSold)
	go appc.dbc.GetCurrentEnteredPaying(cEnteredPaying)

	numCurrentInside, numCurrentSold := <-cEntered, <-cSold
	numCurrentPaying := <-cEnteredPaying

	// check that both goroutines provide a meaningful number
	if numCurrentInside > -1 && numCurrentSold > -1 {
		_, _ = ctx.JSON(iris.Map{
			"status":               iris.StatusOK,
			"currentInside":        numCurrentInside,
			"currentSold":          numCurrentSold,
			"currentPayingEntered": numCurrentPaying,
		})
	} else {
		msg := fmt.Sprintf(
			"error retrieving tickets stats - current inside: %v,"+
				"current sold: %v, currently paying inside: %v",
			numCurrentInside, numCurrentSold, numCurrentPaying)
		appc.InternalErrorMessage(ctx, msg)
	}
}

func (appc *AppController) GetTicketDetails(ctx iris.Context) {
	ticketNum, err := ctx.Params().GetUint("ticketNum")

	if err != nil || ticketNum > dbconn.TICKETHIGH {
		msg := fmt.Sprintf("ticket number %v is not in the correct range [%v, %v]",
			ticketNum, dbconn.TICKETLOW, dbconn.TICKETHIGH)
		appc.BadRequestMessage(ctx, msg)
	} else {
		if attendee, err := appc.dbc.TicketDetails(ticketNum); err != nil {
			appc.InternalErrorMessage(ctx,
				fmt.Sprintf("error retrieving the details of ticket %v", ticketNum))
		} else {
			_, _ = ctx.JSON(iris.Map{
				"status":   iris.StatusOK,
				"attendee": attendee,
			})
		}
	}
}

func (appc *AppController) SetEntered(ctx iris.Context) {
	var ticket dbconn.Ticket
	if err := appc.readPostTicket(ctx, &ticket); err != nil {
		ctx.ResponseWriter().FlushResponse()
		ctx.ResponseWriter().EndResponse()
		return
	}

	// Default result -> the ticket has not been sold
	// and therefore it cannot enter without first pay the entrance
	result := iris.Map{
		"ticketNum": ticket.TicketNum,
		"status":    iris.StatusBadRequest,
		"entered":   false,
		"msg":       "ticket unsold - can not enter",
	}

	status, err := appc.dbc.IsSoldEntered(ticket.TicketNum)

	if err != nil {
		appc.InternalErrorMessage(ctx, fmt.Sprintf(
			"Error encountered while allowing the entrance to ticket %v",
			ticket.TicketNum))
	} else {
		switch status {
		case dbconn.SOLDENTERED:
			ctx.StatusCode(iris.StatusBadRequest)
			// notify that the ticket is already entered,
			// so the same ticket number cannot enter again
			result["entered"] = true
			result["msg"] = "ticket is sold and already entered"
		case dbconn.SOLD:
			if err := appc.dbc.SetEntered(ticket.TicketNum); err != nil {
				appc.InternalErrorMessage(ctx,
					fmt.Sprintf("error occurred setting ticket %v as entered", ticket.TicketNum))
			} else {
				// successful update
				result["status"] = iris.StatusOK
				result["entered"] = true
				result["msg"] = "ticket set as entered correctly"

				appc.logger.Println(fmt.Sprintf("ticket %v entered", ticket.TicketNum))
			}
		case dbconn.UNSOLD:
			ctx.StatusCode(iris.StatusBadRequest)
		}

		_, _ = ctx.JSON(result)
	}

}

func (appc *AppController) SellTicket(ctx iris.Context) {
	var ticket dbconn.Ticket
	if err := appc.readPostTicket(ctx, &ticket); err != nil {
		ctx.ResponseWriter().FlushResponse()
		ctx.ResponseWriter().EndResponse()
		return
	}

	result := iris.Map{
		"ticketNum": ticket.TicketNum,
		"status":    iris.StatusInternalServerError,
		"soldNow":   false,
		"entered":   false,
		"msg":       "failed to sold selected ticket",
	}

	if ticket.FirstName == "" || ticket.LastName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		result["status"] = iris.StatusBadRequest
		result["msg"] = fmt.Sprintf(
			"missing some atteendee details - first_name: %v, last_name: %v",
			ticket.FirstName, ticket.LastName)
	} else {
		// save in the database only capitalized names
		ticket.FirstName = strings.Title(ticket.FirstName)
		ticket.LastName = strings.Title(ticket.LastName)

		// verify whether the ticket can be sold
		if attendee, err := appc.dbc.TicketDetails(ticket.TicketNum); err != nil {
			appc.InternalErrorMessage(ctx, err.Error())
		} else {
			if attendee.FirstName.Valid || attendee.Sold {
				ctx.StatusCode(iris.StatusBadRequest)
				_, _ = ctx.JSON(iris.Map{
					"ticketNum": ticket.TicketNum,
					"status":    iris.StatusBadRequest,
					"soldNow":   false,
					"entered":   attendee.Entered.Valid,
					"msg":       "this ticket can not sold - either reserved or already sold",
				})
				// do not execute the rollback if it is already set as not entered
				ctx.ResponseWriter().FlushResponse()
				ctx.ResponseWriter().EndResponse()
				return
			}
		}

		err := appc.dbc.SellTicket(ticket.TicketNum, ticket.FirstName, ticket.LastName)

		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
		} else {
			result["status"] = iris.StatusOK
			result["soldNow"] = true
			result["entered"] = true
			result["msg"] = fmt.Sprintf("ticket sold correctly to %v %v",
				ticket.FirstName, ticket.LastName)
		}
	}

	_, _ = ctx.JSON(result)
}

func (appc *AppController) ResetTicket(ctx iris.Context) {
	var ticket dbconn.Ticket
	if err := appc.readPostTicket(ctx, &ticket); err != nil {
		ctx.ResponseWriter().FlushResponse()
		ctx.ResponseWriter().EndResponse()
		return
	}

	result := iris.Map{
		"ticketNum": ticket.TicketNum,
		"status":    iris.StatusInternalServerError,
		"msg":       "failed to reset selected ticket",
	}

	if err := appc.dbc.ResetTicket(ticket.TicketNum); err != nil {
		appc.InternalErrorMessage(ctx, err.Error())
	} else {
		result["status"] = iris.StatusOK
		result["msg"] = "ticket reset correctly"
	}

	_, _ = ctx.JSON(result)
}

func (appc *AppController) RollbackEntrance(ctx iris.Context) {
	var ticket dbconn.Ticket
	if err := appc.readPostTicket(ctx, &ticket); err != nil {
		ctx.ResponseWriter().FlushResponse()
		ctx.ResponseWriter().EndResponse()
		return
	}

	// verify whether the ticket is already entered
	if enteredTime, err := appc.dbc.WhenEntered(ticket.TicketNum); err != nil {
		appc.InternalErrorMessage(ctx, err.Error())
	} else {
		if !enteredTime.Valid {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.JSON(iris.Map{
				"ticketNum": ticket.TicketNum,
				"status":    iris.StatusBadRequest,
				"rollback":  false,
				"msg":       "ticket not entered - rollback not performed",
			})
			// do not execute the rollback if it is already set as not entered
			ctx.ResponseWriter().FlushResponse()
			ctx.ResponseWriter().EndResponse()
			return
		}
	}

	if err := appc.dbc.RollbackEntrance(ticket.TicketNum); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.JSON(iris.Map{
			"ticketNum": ticket.TicketNum,
			"status":    iris.StatusInternalServerError,
			"rollback":  false,
			"msg":       "entrance rollback failed",
		})
	} else {
		_, _ = ctx.JSON(iris.Map{
			"ticketNum": ticket.TicketNum,
			"status":    iris.StatusOK,
			"rollback":  true,
			"msg":       "entrance rollback correctly executed",
		})
	}
}

func (appc *AppController) GetTicketVendor(ctx iris.Context) {
	ticketNum, err := ctx.Params().GetUint("ticketNum")

	if err != nil || ticketNum > dbconn.TICKETHIGH {
		msg := fmt.Sprintf("ticket number %v is not in the correct range [%v, %v]",
			ticketNum, dbconn.TICKETLOW, dbconn.TICKETHIGH)
		appc.BadRequestMessage(ctx, msg)
	} else {
		if attendee, err := appc.dbc.TicketVendor(ticketNum); err != nil {
			appc.InternalErrorMessage(ctx,
				fmt.Sprintf("error retrieving the details of ticket %v", ticketNum))
		} else {
			_, _ = ctx.JSON(iris.Map{
				"status": iris.StatusOK,
				"vendor": attendee.Vendor,
			})
		}
	}
}

/* ======= ERROR MANAGEMENT ======= */
func (appc *AppController) BadRequest(ctx iris.Context) {
	appc.BadRequestMessage(ctx, "Provided request can not be understood.")
}
func (appc *AppController) BadRequestMessage(ctx iris.Context, msg string) {
	appc.logger.Println(ctx.Path() + " - " + msg)
	ctx.StatusCode(iris.StatusBadRequest)
	_, _ = ctx.JSON(iris.Map{
		"status": iris.StatusBadRequest,
		"msg":    msg,
	})
}

func (appc *AppController) Unauthorized(ctx iris.Context) {
	appc.UnauthorizedMessage(ctx, "You are not authorized to access this resource.")
}
func (appc *AppController) UnauthorizedMessage(ctx iris.Context, msg string) {
	appc.logger.Println(ctx.Path() + " - " + msg)
	ctx.StatusCode(iris.StatusUnauthorized)
	_, _ = ctx.JSON(iris.Map{
		"status": iris.StatusUnauthorized,
		"msg":    msg,
	})
}

func (appc *AppController) NotFound(ctx iris.Context) {
	appc.NotFoundMessage(ctx,
		fmt.Sprintf("Resource requested at %v not found.", ctx.Path()))
}
func (appc *AppController) NotFoundMessage(ctx iris.Context, msg string) {
	appc.logger.Println(ctx.Path() + " - " + msg)
	ctx.StatusCode(iris.StatusNotFound)
	_, _ = ctx.JSON(iris.Map{
		"status": iris.StatusNotFound,
		"msg":    msg,
	})
}

func (appc *AppController) InternalError(ctx iris.Context) {
	appc.InternalErrorMessage(ctx,
		"Server has encountered an error and can not satisfy the request.")
}
func (appc *AppController) InternalErrorMessage(ctx iris.Context, msg string) {
	appc.logger.Println(ctx.Path() + " - " + msg)
	ctx.StatusCode(iris.StatusInternalServerError)
	_, _ = ctx.JSON(iris.Map{
		"status": iris.StatusInternalServerError,
		"msg":    msg,
	})
}
