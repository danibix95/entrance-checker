/* eslint-disable camelcase */
/*
 * Daniele Bissoli
 * FdP Tickets Manager - Database management
 * v0.0.2 - 2018-03-31
 */

'use strict'

const pg = require('pg')
const Pool = require('pg-pool')
const url = require('url')
const crypto = require('crypto')

const pino = require('pino')
const logger = pino({ level: process.env.LOG_LEVEL ?? 'info' })

/* =========== QUERIES ========= */
// selections
const getCredentials = `SELECT username, password FROM fdp_staff WHERE username = $1::text`
const getAttendees = `SELECT ticket_num, sold, entered FROM attendees ORDER BY ticket_num`
const getAttendee = `SELECT ticket_num, first_name, last_name, ticket_type, sold, entered FROM attendees WHERE ticket_num = $1::integer`
// const checkEntered = `SELECT entered FROM attendees WHERE ticket_num = $1::integer AND entered IS NULL`;
const checkEntered = `SELECT sold FROM attendees WHERE ticket_num = $1::integer AND entered IS NULL`
const checkEnteredNum = `SELECT COUNT(*) FROM attendees WHERE entered IS NOT NULL`
const checkSoldNum = `SELECT COUNT(*) FROM attendees WHERE sold = TRUE`
const getEntrance = `SELECT entered FROM attendees WHERE ticket_num = $1::integer`
const getVendor = `SELECT vendor FROM attendees WHERE ticket_num = $1::integer`

const updateStatus = `UPDATE attendees SET entered = NOW() WHERE ticket_num = $1::integer`
const sellTicket = `UPDATE attendees SET last_name = $3::text, first_name = $2::text, sold = true, resp_vendor = 'entrance', entered = NOW() WHERE ticket_num = $1::integer`
const uncommitEntrance = `UPDATE attendees SET entered = NULL WHERE ticket_num = $1::integer`

/* ============================= */

/* ========= UTILITIES ========= */
/**
 * Gets the hash of the input message.
 *
 * @param      {string}  message  The message
 * @return     {string}  The hash.
 */
function getHash(message) {
  return crypto.createHash('sha256')
    .update(message)
    .digest('base64')
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
      logger.log('error', error)
      return false
    })
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
      logger?.error(error)
      return false
    })
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
      logger?.error(error)
      return false
    })
}

class DB {
  constructor() {
    // !! REMEMBER !! Here you need to check that env variable was set before run!
    if (process.env.DATABASE_URL === undefined) {
      logger.fatal("\nYou'd not set DATABASE_URL variable. Check instructions before (see data folder)!\n")
    }
    const pgParams = url.parse(process.env.DATABASE_URL)
    const pgAuth = pgParams.auth.split(':')

    // database connection configuration
    const config = {
      user: pgAuth[0],
      database: pgParams.pathname.split('/')[1],
      password: pgAuth[1],
      host: pgParams.hostname,
      port: pgParams.port,
      ssl: false, // require ssl connection
      max: 10, // max number of clients in the pool
      idleTimeoutMillis: 30000, // how long a client is allowed to remain idle before being closed
    }

    // connection pool for database
    this.pool = new Pool(config)

    // manage connection pool error
    this.pool.on('error', (error, client) => {
      logger.error(error)
    })
  }

  login(username, password) {
    // check if password is a string
    if (typeof password !== 'string') { throw new Error('Wrong kind of password') }

    const providedPwd = getHash(password)

    return makeSelection(this.pool, getCredentials, [username], logger)
      .then((result) => {
        // if rowCount = 1 means that
        // exist one and only one user
        // with that username (provided by primary key on username)
        if (result.length === 1) {
          return (result[0].password === providedPwd) ? 0 : 1
        } else if (result.rowCount === 0) {
          // code representing that no user is found
          return 2
        }

        // notify who use this to send an error 500
        return 3
      })
      .catch((error) => { logger.error(error) })
  }

  /** Give the entrance date of specified ticket */
  entered(ticket_num) {
    return makeSelection(this.pool, getEntrance, [ticket_num], logger)
      .then((result) => {
        if (result.length === 1) {
          return result[0].entered
        }

        throw new Error('Error getting entrance date')
      })
      .catch((error) => { logger.log('error', error) })
  }

  /** Check if specified ticket, then in case of success update DB saving the date */
  setEntered1(ticket_num) {
    //  -1 => update error
    //   0 => OK
    //   1 => ticket unsold
    //   2 => ticket already entered
    return makeSelection(this.pool, checkEntered, [ticket_num], logger)
      .then((result) => {
        if (result.length === 1) {
          return (result[0].sold ? 0 : 1)
        }
        // attendee is already entered
        return 2
      })
      .catch((error) => { logger.log(error) })
  }

  setEntered2(ticket_num) {
    return makeUpdate(this.pool, updateStatus, [ticket_num])
      .then((result) => result === 1)
      .catch((error) => { logger.log(error) })
  }

  list() {
    return makeSelection(this.pool, getAttendees, [], logger)
      .then((result) => result)
      .catch((error) => { logger.log(error) })
  }

  currentInside() {
    return makeSelection(this.pool, checkEnteredNum, [], logger)
      .then((result) => result[0].count)
      .catch((error) => { logger.error(error) })
  }

  currentSold() {
    return makeSelection(this.pool, checkSoldNum, [], logger)
      .then((result) => result[0].count)
      .catch((error) => { logger.error(error) })
  }

  details(tnum) {
    return makeSelection(this.pool, getAttendee, [tnum], logger)
      .then((result) => result[0])
      .catch((error) => { logger.error(error) })
  }

  sell(tnum, fname, lname) {
    return makeSelection(this.pool, getAttendee, [tnum], logger)
      .then((result) => result[0])
      .then((ticket) => {
        // if it was already sold, then I can't sell it again
        if (ticket.sold) {
          return 2
        }

        return makeUpdate(this.pool, sellTicket, [tnum, fname, lname], logger)
      })
      .then((result) => {
        if (result === 1) { return 1 } /* ticket sold */
        if (result === 2) { return 2 } /* ticket already sold, so no update */
        return 0 /* some error around */
      })
      .catch((error) => { logger.error(error) })
  }

  deleteEntrance(tnum) {
    return makeUpdate(this.pool, uncommitEntrance, [tnum], logger)
      .then((result) => result === 1)
      .catch((error) => { logger.error(error) })
  }

  checkVendor(tnum) {
    return makeSelection(this.pool, getVendor, [tnum], logger)
      .then((result) => {
        if (result.length === 1) {
          return result[0].vendor
        }

        return undefined
      })
      .catch((error) => { logger.error(error) })
  }
}

module.exports = DB
