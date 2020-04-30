package dbconn

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"io"
	"log"
	"os"
)

/* =========== QUERIES ========= */
const getCredentials = `SELECT username, password FROM fdp_staff WHERE username = $1::text`
const getAttendees = `SELECT ticket_num, ticket_type, sold, entered FROM attendees ORDER BY ticket_num`
const getAttendee = `SELECT ticket_num, first_name, last_name, ticket_type, sold, vendor` +
	`, resp_vendor, entered FROM attendees WHERE ticket_num = $1::integer`

const isSoldEntered = `SELECT sold, entered FROM attendees WHERE ticket_num = $1::integer`
const checkEnteredNum = `SELECT COUNT(*) FROM attendees WHERE entered IS NOT NULL`
const checkSoldNum = `SELECT COUNT(*) FROM attendees WHERE sold = TRUE`
const checkSoldEnteredNum = `SELECT COUNT(*) FROM attendees WHERE entered IS NOT NULL ` +
	`AND sold = TRUE AND ticket_type > 0`
const getEntrance = `SELECT entered FROM attendees WHERE ticket_num = $1::integer`
const getVendor = `SELECT vendor FROM attendees WHERE ticket_num = $1::integer`

const updateStatus = `UPDATE attendees SET entered = NOW() WHERE ticket_num = $1::integer`
const sellTicket = `UPDATE attendees SET last_name = $3::text, first_name = $2::text,` +
	`sold = true, resp_vendor = 'Entrance', entered = NOW() WHERE ticket_num = $1::integer`
const resetTicket = `UPDATE attendees SET last_name = NULL, first_name = NULL,` +
	`sold = false, vendor = NULL, entered = NULL WHERE ticket_num = $1::integer`
const rollbackEntrance = `UPDATE attendees SET entered = NULL WHERE ticket_num = $1::integer`

var statements map[string]*sql.Stmt

/* ======= TYPES DEFINITION ====== */
type EnteredStatus uint8

const (
	SOLDENTERED EnteredStatus = iota + 1
	SOLD
	UNSOLD
)

type DBController struct {
	db     *sql.DB
	logger *log.Logger
}

func New(logFile *os.File) *DBController {
	var dbc DBController

	dbc.logger = log.New(io.MultiWriter(logFile), "", log.LstdFlags)
	dbc.connect()
	dbc.prepareQueries()

	return &dbc
}

// Initialize the database controller and prepare relevant queries to be executed
func (dbc *DBController) connect() {
	// read the information for connecting to postgres database from postgres_info file
	dbConf, envErr := godotenv.Read("postgres_info")
	if envErr != nil {
		dbc.logger.Fatal("Error loading connection info from specified file!")
	}
	var psqlInfo = fmt.Sprintf(
		"host=%s port=%v user=%s password=%s dbname=%s sslmode=disable",
		dbConf["HOST"], dbConf["PORT"], dbConf["USER"], dbConf["PWD"], dbConf["DB_NAME"])

	var dbErr error
	dbc.db, dbErr = sql.Open("postgres", psqlInfo)
	if dbErr != nil {
		dbc.logger.Fatal(fmt.Sprintf("Impossible to connect to the database! %v", dbErr))
	}

	// check that the connection works properly
	dbc.PingDB()
	dbc.logger.Println("Connection established!")
}

func (dbc *DBController) PingDB() {
	pingErr := dbc.db.Ping()
	if pingErr != nil {
		dbc.logger.Panic(pingErr)
	} else {
		dbc.logger.Println("Database reachable!")
	}
}

// Prepare relevant queries for the app
func (dbc *DBController) prepareQueries() {
	stmtMap := map[string]*sql.Stmt{}

	stmtMap["getCredential"], _ = dbc.db.Prepare(getCredentials)
	stmtMap["getAttendees"], _ = dbc.db.Prepare(getAttendees)
	stmtMap["getAttendee"], _ = dbc.db.Prepare(getAttendee)
	stmtMap["isSoldEntered"], _ = dbc.db.Prepare(isSoldEntered)
	stmtMap["checkEnteredNum"], _ = dbc.db.Prepare(checkEnteredNum)
	stmtMap["checkSoldNum"], _ = dbc.db.Prepare(checkSoldNum)
	stmtMap["checkSoldEnteredNum"], _ = dbc.db.Prepare(checkSoldEnteredNum)
	stmtMap["getEntrance"], _ = dbc.db.Prepare(getEntrance)
	stmtMap["getVendor"], _ = dbc.db.Prepare(getVendor)
	stmtMap["updateStatus"], _ = dbc.db.Prepare(updateStatus)
	stmtMap["sellTicket"], _ = dbc.db.Prepare(sellTicket)
	stmtMap["resetTicket"], _ = dbc.db.Prepare(resetTicket)
	stmtMap["rollbackEntrance"], _ = dbc.db.Prepare(rollbackEntrance)

	statements = stmtMap
}

func (dbc *DBController) CloseDB() {
	err := dbc.db.Close()
	if err != nil {
		dbc.logger.Panicln("An error occurred closing the database connection!")
	}
}

/* ============= DATABASE ACCESSES ============= */
func getHash(message string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// Retrieve from the database the timestamp of when selected tickets checked-in at Party's entrance
func (dbc *DBController) WhenEntered(ticketNum uint) (pq.NullTime, error) {
	attendee := Attendee{TicketNum: ticketNum}

	var userErr error
	err := statements["getEntrance"].QueryRow(attendee.TicketNum).
		Scan(&attendee.Entered)
	if err != nil {
		userErr = fmt.Errorf("impossible to retrieve when ticket %v has entered",
			attendee.TicketNum)
		dbc.logger.Println(userErr.Error(), err)
	}

	return attendee.Entered, userErr
}

// Check among the attendees that has not entered,
// whether selected tickets has already been sold
func (dbc *DBController) IsSoldEntered(ticketNum uint) (EnteredStatus, error) {
	attendee := Attendee{TicketNum: ticketNum}

	var userErr error
	err := statements["isSoldEntered"].QueryRow(attendee.TicketNum).
		Scan(&attendee.Sold, &attendee.Entered)
	if err != nil {
		if err == sql.ErrNoRows {
			userErr = fmt.Errorf("ticket %v does not exist (not in the valid range)",
				attendee.TicketNum)
		} else {
			userErr = fmt.Errorf("impossible to check whether ticket %v has been sold",
				attendee.TicketNum)
		}
		dbc.logger.Println(userErr.Error(), err)
	}

	status := SOLDENTERED
	if !attendee.Sold {
		status = UNSOLD
	} else {
		if !attendee.Entered.Valid {
			status = SOLD
		}
	}

	return status, userErr
}

// Set specified ticket as entered (change entered value to current datetime)
func (dbc *DBController) SetEntered(ticketNum uint) error {
	var userErr error
	attendee := Attendee{TicketNum: ticketNum}

	res, err := statements["updateStatus"].Exec(attendee.TicketNum)
	if err != nil {
		userErr = fmt.Errorf("impossible to set as entered ticket %v",
			attendee.TicketNum)
		dbc.logger.Println(userErr.Error(), err)
	} else {
		count, err := res.RowsAffected()
		if err != nil || count != 1 {
			userErr = fmt.Errorf("something went wrong updating entrance of ticket %v",
				attendee.TicketNum)
			dbc.logger.Println(userErr.Error(),
				fmt.Sprintf("(changed %v rows)", count), err)
		}
	}

	return userErr
}

// Revert the entrance of selected ticket
func (dbc *DBController) RollbackEntrance(ticketNum uint) error {
	var userErr error
	attendee := Attendee{TicketNum: ticketNum}

	res, err := statements["rollbackEntrance"].Exec(attendee.TicketNum)
	if err != nil {
		userErr = fmt.Errorf("impossible to set reset ticket %v entrance",
			attendee.TicketNum)
		dbc.logger.Println(userErr.Error(), err)
	} else {
		count, err := res.RowsAffected()
		if err != nil || count != 1 {
			userErr = fmt.Errorf("something went wrong updating entrance of ticket %v",
				attendee.TicketNum)
			dbc.logger.Println(userErr.Error(),
				fmt.Sprintf("(changed %v rows)", count), err)
		}
	}

	return userErr
}

// Return the information of the specified ticket
func (dbc *DBController) TicketDetails(ticketNum uint) (Attendee, error) {
	var attendee = Attendee{TicketNum: ticketNum}
	var userErr error

	err := statements["getAttendee"].QueryRow(attendee.TicketNum).
		Scan(&attendee.TicketNum, &attendee.FirstName, &attendee.LastName,
			&attendee.TicketType, &attendee.Sold, &attendee.Vendor,
			&attendee.RespVendor, &attendee.Entered)

	if err != nil {
		if err == sql.ErrNoRows {
			userErr = fmt.Errorf("ticket %v does not exist (not in the valid range)",
				attendee.TicketNum)
		} else {
			userErr = fmt.Errorf("impossible to retrieve ticket %v details",
				attendee.TicketNum)
		}
		dbc.logger.Println(userErr.Error(), err)
	}

	return attendee, userErr
}

// Return the information of the specified ticket
func (dbc *DBController) TicketVendor(ticketNum uint) (Attendee, error) {
	var attendee = Attendee{TicketNum: ticketNum}
	var userErr error

	err := statements["getVendor"].QueryRow(attendee.TicketNum).Scan(&attendee.Vendor)

	if err != nil {
		if err == sql.ErrNoRows {
			userErr = fmt.Errorf("ticket %v does not exist (not in the valid range)",
				attendee.TicketNum)
		} else {
			userErr = fmt.Errorf("impossible to retrieve ticket %v details, %v",
				attendee.TicketNum, err.Error())
		}
		dbc.logger.Println(userErr.Error(), err)
	}

	return attendee, userErr
}

// Return the list of all the tickets
func (dbc *DBController) TicketsList() ([]AttendeeSimple, error) {
	var attendees []AttendeeSimple
	var userErr error
	rows, err := statements["getAttendees"].Query()
	defer rows.Close()

	if err != nil {
		userErr = errors.New("impossible to retrieve the list of tickets")
		dbc.logger.Println(userErr.Error(), err)
	} else {
		numErrors := 0
		for rows.Next() {
			var attendee AttendeeSimple
			err = rows.Scan(&attendee.TicketNum, &attendee.TicketType,
				&attendee.Sold, &attendee.Entered)

			if err != nil {
				numErrors++
				dbc.logger.Println(fmt.Sprintf("impossible to read ticket %v"+
					" and insert it into the attendees list", attendee.TicketNum), err)
			} else {
				attendees = append(attendees, attendee)
			}
		}
		if numErrors > 0 {
			userErr = fmt.Errorf("%v attendees tickets were not loaded correctly",
				numErrors)
			dbc.logger.Println(userErr.Error())
		}
	}

	return attendees, userErr
}

// Retrieve the amount of attendees that are within Collegio's perimeter
func (dbc *DBController) GetCurrentInside(result chan int) {
	numPeopleEntered := 0
	err := statements["checkEnteredNum"].QueryRow().Scan(&numPeopleEntered)

	if err != nil {
		dbc.logger.Println("error retrieving the number of people entered")
		result <- -1 // notify the error through the channel
	} else {
		result <- numPeopleEntered
	}
}

// Retrieve the amount of tickets that have been currently sold (and paid)
func (dbc *DBController) GetCurrentSold(result chan int) {
	numTicketSold := 0
	err := statements["checkSoldNum"].QueryRow().Scan(&numTicketSold)

	if err != nil {
		dbc.logger.Println("error retrieving the number of tickets sold")
		result <- -1 // notify the error through the channel
	} else {
		result <- numTicketSold
	}
}

// Retrieve the amount of tickets that have been currently sold (paid) and entered
// This function does not consider free tickets (omaggi)
func (dbc *DBController) GetCurrentEnteredPaying(result chan int) {
	numTicketSoldEntered := 0
	err := statements["checkSoldEnteredNum"].QueryRow().Scan(&numTicketSoldEntered)

	if err != nil {
		dbc.logger.Println("error retrieving the number of paying tickets entered")
		result <- -1 // notify the error through the channel
	} else {
		result <- numTicketSoldEntered
	}
}

// Sell selected ticket to a specific person
func (dbc *DBController) SellTicket(ticketNum uint, firstName, lastName string) error {
	var userErr error

	res, err := statements["sellTicket"].Exec(ticketNum, firstName, lastName)
	if err != nil {
		userErr = fmt.Errorf(
			"impossible to sell ticket %v", ticketNum)
		dbc.logger.Println(userErr.Error(), err)
	} else {
		count, err := res.RowsAffected()
		if err != nil || count != 1 {
			userErr = fmt.Errorf("something went wrong selling ticket %v", ticketNum)
			dbc.logger.Println(userErr.Error(),
				fmt.Sprintf("(changed %v rows)", count), err)
		}
	}

	return userErr
}

// Sell selected ticket to a specific person
func (dbc *DBController) ResetTicket(ticketNum uint) error {
	var userErr error

	res, err := statements["resetTicket"].Exec(ticketNum)
	if err != nil {
		userErr = fmt.Errorf(
			"impossible to reset ticket %v", ticketNum)
		dbc.logger.Println(userErr.Error(), err)
	} else {
		count, err := res.RowsAffected()
		if err != nil || count != 1 {
			userErr = fmt.Errorf("something went wrong resetting ticket %v", ticketNum)
			dbc.logger.Println(userErr.Error(),
				fmt.Sprintf("(changed %v rows)", count), err)
		}
	}

	return userErr
}
