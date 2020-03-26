package dbconn

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
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
const getAttendees = `SELECT ticket_num, sold, entered FROM attendees ORDER BY ticket_num`
const getAttendee = `SELECT ticket_num, first_name, last_name, ticket_type, sold, entered
						FROM attendees WHERE ticket_num = $1::integer`

const isSoldEntered = `SELECT sold, entered FROM attendees WHERE ticket_num = $1::integer`
const checkEnteredNum = `SELECT COUNT(*) FROM attendees WHERE entered IS NOT NULL`
const checkSoldNum = `SELECT COUNT(*) FROM attendees WHERE sold = TRUE`
const getEntrance = `SELECT entered FROM attendees WHERE ticket_num = $1::integer`
const getVendor = `SELECT vendor FROM attendees WHERE ticket_num = $1::integer`

const updateStatus = `UPDATE attendees SET entered = NOW() WHERE ticket_num = $1::integer`
const sellTicket = `UPDATE attendees SET last_name = $3::text, first_name = $2::text,` +
	`sold = true, resp_vendor = 'entrance', entered = NOW() WHERE ticket_num = $1::integer`
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

	dbc.logger = log.New(io.MultiWriter(os.Stderr, logFile),
		"", log.LstdFlags)
	dbc.connect()
	dbc.prepareQueries()

	return &dbc
}

// Initialize the database controller
// and prepare relevant queries to be executed
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
		panic(fmt.Sprintf("Impossible to connect to the database! %v", dbErr))
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
	stmtMap["getEntrance"], _ = dbc.db.Prepare(getEntrance)
	stmtMap["getVendor"], _ = dbc.db.Prepare(getVendor)
	stmtMap["updateStatus"], _ = dbc.db.Prepare(updateStatus)
	stmtMap["sellTicket"], _ = dbc.db.Prepare(sellTicket)
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
	return base64.StdEncoding.EncodeToString([]byte(hash.Sum(nil)))
}

func (dbc *DBController) WhenEntered(ticketNum uint) pq.NullTime {
	attendee := Attendee{TicketNum: ticketNum}

	err := statements["getEntrance"].QueryRow(attendee.TicketNum).
		Scan(&attendee.Entered)
	if err != nil {
		dbc.logger.Panicln(fmt.Sprintf(
			"Impossible to retrieve when ticket %v has entered!",
			attendee.TicketNum))
	}

	return attendee.Entered
}

// Check among the attendees that has not entered,
// whether selected tickets has already been sold
func (dbc *DBController) IsSoldEntered(ticketNum uint) EnteredStatus {
	attendee := Attendee{TicketNum: ticketNum}

	err := statements["isSoldEntered"].QueryRow(attendee.TicketNum).
		Scan(&attendee.Sold, &attendee.Entered)
	if err != nil {
		if err == sql.ErrNoRows {
			dbc.logger.Println(err)
			dbc.logger.Panicln(fmt.Sprintf(
				"Ticket %v does not exist (not in the valid range)!",
				attendee.TicketNum))
		} else {
			dbc.logger.Panicln(fmt.Sprintf(
				"Impossible to check whether ticket %v has been sold!\n%v",
				attendee.TicketNum, err))
		}
	}

	status := SOLDENTERED
	if !attendee.Sold {
		status = UNSOLD
	} else {
		if !attendee.Entered.Valid {
			status = SOLD
		}
	}

	return status
}

// Set specified ticket as entered (change entered value to current datetime)
func (dbc *DBController) SetEntered(ticketNum uint) bool {
	result := true
	attendee := Attendee{TicketNum: ticketNum}

	res, err := statements["updateStatus"].Exec(attendee.TicketNum)
	if err != nil {
		result = false
		dbc.logger.Panicln(fmt.Sprintf(
			"Impossible to set as entered ticket %v!",
			attendee.TicketNum))
	}
	count, err := res.RowsAffected()
	if err != nil || count != 1 {
		result = false
		dbc.logger.Panicln(fmt.Sprintf(
			"Something went wrong updating entrance of ticket %v! (changed %v rows)",
			attendee.TicketNum, count))
	}

	return result
}

// Revert the entrance of selected ticket
func (dbc *DBController) RollbackEntrance(ticketNum uint) bool {
	result := true
	attendee := Attendee{TicketNum: ticketNum}

	res, err := statements["rollbackEntrance"].Exec(attendee.TicketNum)
	if err != nil {
		result = false
		dbc.logger.Panicln(fmt.Sprintf(
			"Impossible to set reset ticket %v entrance!",
			attendee.TicketNum))
	}
	count, err := res.RowsAffected()
	if err != nil || count != 1 {
		result = false
		dbc.logger.Panicln(fmt.Sprintf(
			"Something went wrong updating entrance of ticket %v! (changed %v rows)",
			attendee.TicketNum, count))
	}

	return result
}

// Return the information of the specified ticket
func (dbc DBController) TicketDetails(ticketNum uint) (*Attendee, error) {
	var attendee = Attendee{TicketNum: ticketNum}

	err := statements["getAttendee"].QueryRow(attendee.TicketNum).
		Scan(&attendee.TicketNum, &attendee.FirstName, &attendee.LastName,
			&attendee.TicketType, &attendee.Sold, &attendee.Entered)
	if err != nil {
		if err == sql.ErrNoRows {
			dbc.logger.Println(err)
			dbc.logger.Panicln(fmt.Sprintf(
				"Ticket %v does not exist (not in the valid range)!",
				attendee.TicketNum))
		}
		// let upper level manage potential error
		return nil, err
	}

	return &attendee, err
}

// Return the list of all the tickets
func (dbc *DBController) TicketsList() ([]Attendee, error) {
	var attendees []Attendee
	rows, err := statements["getAttendees"].Query()

	if err != nil {
		dbc.logger.Panicln("Impossible to retrieve the list of tickets!")
	}

	for rows.Next() {
		var attendee Attendee
		err := rows.Scan(&attendee.TicketNum, &attendee.FirstName,
			&attendee.LastName, &attendee.TicketType,
			&attendee.Sold, &attendee.Entered)
		if err != nil {
			dbc.logger.Println(fmt.Sprintf("Impossible to read ticket %v"+
				"and insert into the attendees list!", attendee.TicketNum))
		} else {
			attendees = append(attendees, attendee)
		}
	}

	return attendees, err
}
