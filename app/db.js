/*
 * Daniele Bissoli
 * FdP Tickets Manager - Database management
 * v0.0.1 - 2017-04-16
 */

const pg = require("pg");
const Pool = require("pg-pool");
const url = require("url");
const crypto = require('crypto');
// application library
const logger = require("./logger.js");

/* =========== QUERIES ========= */
// selections
const getCredentials = `SELECT username, password FROM fdp_staff WHERE username = $1::text`;
const getAttendees = `SELECT ticket_num, sold, entered FROM attendees ORDER BY ticket_num`;
const getAttendee = `SELECT ticket_num, first_name, last_name, ticket_type, sold, entered FROM attendees WHERE ticket_num = $1::integer`;
//const checkEntered = `SELECT entered FROM attendees WHERE ticket_num = $1::integer AND entered IS NULL`;
const checkEntered = `SELECT sold FROM attendees WHERE ticket_num = $1::integer AND entered IS NULL`;
const checkEnteredNum = `SELECT COUNT(*) FROM attendees WHERE entered IS NOT NULL`;
const getEntrance = `SELECT entered FROM attendees WHERE ticket_num = $1::integer`;
const getVendor = `SELECT vendor FROM attendees WHERE ticket_num = $1::integer`;

const updateStatus = `UPDATE attendees SET entered = NOW() WHERE ticket_num = $1::integer`;
const sellTicket = `UPDATE attendees SET last_name = $3::text, first_name = $2::text, sold = true, vendor = 'entrance', entered = NOW() WHERE ticket_num = $1::integer`;
const uncommitEntrance = `UPDATE attendees SET entered = NULL WHERE ticket_num = $1::integer`;
/* ============================= */

/* ========= UTILITIES ========= */
/**
 * Gets the hash of the input message.
 *
 * @param      {string}  message  The message
 * @return     {string}  The hash.
 */
function getHash(message) {
  return crypto.createHash("sha256")
      .update(message)
      .digest("base64");
}

/**
 * Makes an insertion on db.
 *
 * @param      {string}  query   The query
 * @param      {array}    params  The parameters
 * @return     {bool}    true if value is inserted correctly, otherwise false
 */
function makeInsertion(pool, query, params) {
  return pool.query(query, params)
      .then((result) => (result.rowCount === 1))
      .catch((error) => {
        logger.log("error", error);
        return false;
      });
}

/**
 * Makes a selection on the db.
 *
 * @param      {string}   query   The query
 * @param      {Array}    params  The parameters
 * @return     {Array || bool} the query result as array of row, otherwise false
 */
function makeSelection(pool, query, params) {
  return pool.query(query, params)
      // return an array of row as query result
      .then((result) => result.rows)
      .catch((error) => {
        logger.log("error", error);
        return false;
      });
}

/**
 * Makes an update on the db.
 *
 * @param      {string}   query   The query
 * @param      {Array}    params  The parameters
 * @return     {Array || bool} the query result as array of row, otherwise false
 */
function makeUpdate(pool, query, params) {
  return pool.query(query, params)
  // return an array of row as query result
      .then((result) => result.rowCount)
      .catch((error) => {
        logger.log("error", error);
        return false;
      });
}

class DB {
  constructor() {
    const pgParams = url.parse(process.env.DATABASE_URL);
    const pgAuth = pgParams.auth.split(":");

    // database connection configuration
    const config = {
      user: pgAuth[0],
      database: pgParams.pathname.split("/")[1],
      password: pgAuth[1],
      host: pgParams.hostname,
      port: pgParams.port,
      ssl: false,  // require ssl connection
      max: 10,    // max number of clients in the pool
      idleTimeoutMillis: 30000, // how long a client is allowed to remain idle before being closed
    };

    // connection pool for database
    this.pool = new Pool(config);

    // manage connection pool error
    this.pool.on('error', function (err, client) {
      logger.log("error", err.message + "\n" + err.stack);
    });
  }

  login(username, password) {
    // check if password is a string
    if (typeof password !== "string") throw new Error("Wrong kind of password");

    const providedPwd = getHash(password);

    return makeSelection(this.pool, getCredentials, [username])
      .then((result) => {
        // if rowCount = 1 means that
        // exist one and only one user
        // with that username (provided by primary key on username)
        if (result.length === 1) {
          return (result[0].password === providedPwd) ? 0 : 1;
        }
        else if (result.rowCount == 0) {
          // code representing that no user is found
          return 2;
        }
        else {
          // notify who use this to send an error 500
          return 3;
        }
      })
      .catch((error) => { logger.log("error", error); });
  }

  /** Give the entrance date of specified ticket */
  entered(ticket_num) {
    return makeSelection(this.pool, getEntrance, [ticket_num])
        .then((result) => {
          if(result.length === 1) {
            return result[0].entered;
          }
          else {
            throw new Error("Error getting entrance date");
          }
        })
        .catch((error) => { logger.log("error", error); });
  }

  /** Check if specified ticket, then in case of success update DB saving the date */
  setEntered1(ticket_num) {
    /*
       -1 => update error
        0 => OK
        1 => ticket unsold
        2 => ticket already entered
    */
    return makeSelection(this.pool, checkEntered, [ticket_num])
        .then((result) => {
          if (result.length === 1) {
            return (result[0].sold ? 0 : 1);
          }
          // attendee is already entered
          return 2;
        })
        .catch((error) => { logger.log("error", error); });
  }

  setEntered2(ticket_num) {
    return makeUpdate(this.pool, updateStatus, [ticket_num])
        .then((result) => result === 1)
        .catch((error) => { logger.log("error", error); });
  }

  list() {
    return makeSelection(this.pool, getAttendees, [])
        .then((result) => result)
        .catch((error) => {logger.log("error", error); });
  }

  currentInside() {
    return makeSelection(this.pool, checkEnteredNum, [])
        .then((result) => result[0].count)
        .catch((error) => { logger.log("error", error); });
  }

  details(tnum) {
    return makeSelection(this.pool, getAttendee, [tnum])
        .then((result) => result[0])
        .catch((error) => { logger.log("error", error); });
  }

  sell(tnum, fname, lname) {
    return makeUpdate(this.pool, sellTicket, [tnum, fname, lname])
        .then((result) => result === 1)
        .catch((error) => { logger.log("error", error); });
  }

  deleteEntrance(tnum) {
    return makeUpdate(this.pool, uncommitEntrance, [tnum])
        .then((result) => result === 1)
        .catch((error) => { logger.log("error", error); });
  }

  checkVendor(tnum) {
    return makeSelection(this.pool, getVendor, [tnum])
        .then((result) => {
          if (result.length === 1) {
            return result[0].vendor;
          }
          else {
            return undefined;
          }
        })
        .catch((error) => { logger.log("error", error); });
  }
}

module.exports = DB;