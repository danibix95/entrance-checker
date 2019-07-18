package dbconn

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"io"
	"log"
	"os"
	"path/filepath"
	_ "path/filepath"
)

/* =========== QUERIES ========= */
const getCredentials = `SELECT username, password FROM fdp_staff WHERE username = $1::text`
const getAttendees = `SELECT ticket_num, sold, entered FROM attendees ORDER BY ticket_num`
const getAttendee = `SELECT ticket_num, first_name, last_name, ticket_type, sold, entered FROM attendees WHERE ticket_num = $1::integer`

const checkEntered = `SELECT sold FROM attendees WHERE ticket_num = $1::integer AND entered IS NULL`
const checkEnteredNum = `SELECT COUNT(*) FROM attendees WHERE entered IS NOT NULL`
const checkSoldNum = `SELECT COUNT(*) FROM attendees WHERE sold = TRUE`
const getEntrance = `SELECT entered FROM attendees WHERE ticket_num = $1::integer`
const getVendor = `SELECT vendor FROM attendees WHERE ticket_num = $1::integer`

const updateStatus = `UPDATE attendees SET entered = NOW() WHERE ticket_num = $1::integer`
const sellTicket = `UPDATE attendees SET last_name = $3::text, first_name = $2::text,` +
	`sold = true, resp_vendor = 'entrance', entered = NOW() WHERE ticket_num = $1::integer`
const rollbackEntrance = `UPDATE attendees SET entered = NULL WHERE ticket_num = $1::integer`

type DBController struct {
	db         *sql.DB
	statements *map[string]*sql.Stmt
	logger     *log.Logger
}

func New(logDir string) *DBController {
	var dbc DBController

	dbc.startLogger(logDir)
	dbc.connect()
	dbc.prepareQueries()

	return &dbc
}

// Define a custom logger to store info also in a log file
func (dbc *DBController) startLogger(logDir string) {
	// to avoid potential errors, create the logs directory if it does not exists
	if _, err := os.Stat(logDir); err != nil {
		if os.IsNotExist(err) {
			dirErr := os.Mkdir(logDir, os.ModeDir)
			if dirErr != nil {
				panic("Impossible to find or create logs folder!")
			}
		}
	}

	// select the file on which the logger will store the information
	logFile, err := os.OpenFile(filepath.Join(logDir, "db_conn.log"),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	dbc.logger = log.New(io.MultiWriter(os.Stderr, logFile), "", log.LstdFlags)
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
	stmtMap["checkEntered"], _ = dbc.db.Prepare(checkEntered)
	stmtMap["checkEnteredNum"], _ = dbc.db.Prepare(checkEnteredNum)
	stmtMap["checkSoldNum"], _ = dbc.db.Prepare(checkSoldNum)
	stmtMap["getEntrance"], _ = dbc.db.Prepare(getEntrance)
	stmtMap["getVendor"], _ = dbc.db.Prepare(getVendor)
	stmtMap["updateStatus"], _ = dbc.db.Prepare(updateStatus)
	stmtMap["sellTicket"], _ = dbc.db.Prepare(sellTicket)
	stmtMap["rollbackEntrance"], _ = dbc.db.Prepare(rollbackEntrance)

	dbc.statements = &stmtMap
}

func (dbc *DBController) CloseDB() {
	err := dbc.db.Close()
	if err != nil {
		panic("An error occurred closing the database connection!")
	}
}
