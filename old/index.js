/*
 * Daniele Bissoli
 * FdP Tickets Manager
 * v0.0.2 - 2017-04-16
 */

'use strict'

/* EXTERNAL LIBRARIES */
const express = require('express')
const path = require('path')
const bodyParser = require('body-parser')
const session = require('express-session')
const compression = require('compression')
const minify = require('express-minify')

const pino = require('pino')
const logger = pino({ level: process.env.LOG_LEVEL ?? 'info' })

const Control = require('./app/control.js')

/* INTERNAL LIBRARIES */
const server = new Control()

/* APPLICATION INITIALIZATION */
// specify process name, for eventually recognize it later
process.title = 'fdp_tickets'
// process.env.NODE_ENV = "production";
// Initialize express application
const app = express()

// set some application information
app.set('title', 'FdP Tickets Manager')
app.set('port', (process.env.PORT || 52017))
// tell to express to use pug engine
// template to render files
app.set('view engine', 'pug')

/* APPLICATION LOGIC */
// enable file compression
app.use(compression())
// enable minifying files
app.use(minify())
// enable application to parse encoded form
app.use(bodyParser.urlencoded({ extended: true }))

// enable session management
app.use(session({
  // TODO: rewrite this management
  // string with hash session cookie
  secret: '946f503c568fdf64d095c9121ae652b3815df5ec95c1a4f11d7b6f0ae180c863',
  // don't save again a session if
  // it isn't modify from previous request
  resave: false,
  // don't save session for user
  // that haven't logged in
  saveUninitialized: false,
  // cookies options
  cookie: {
    sameSite: true,
    // need a certificate to only go over https
    secure: false,
    maxAge: 28800000,
  },
  // set session cookie name
  name: 'nsession',
}))

/* ===================*/
/* MANAGE API */
// these let user to skip login phase
// if it has already logged in
app.get('/', server.isLoggedIn, server.startPage)

// process login request checking user credentials
app.post('/login', server.login)
// process logout request
app.get('/logout', server.logout)
// serve private default page
app.get('/home', server.requireLogin, server.home)

// set everything under /home/ path
// to be protected by unauthorized user
app.get('/home/*', server.requireLogin)

// check that only admin can sell new tickets
app.use('/home/admin/*', server.checkAdmin)

// list tickets status
app.get('/home/tickets', server.getTickets)
// get a notification with tickets info (entered vs sold)
app.get('/home/tickets-info', server.getTicketsInfo)
// get specific ticket status
app.get('/home/tickets/:ticket_num', server.ticketDetails)

// set an user as entered to the event
app.post('/home/tickets/entered', server.entered)
// used to confirm the entrance of attendee
app.post('/home/tickets/entered/commit', server.entered2)

// admin dashboard
app.get('/home/admin/dashboard', server.dashboard)
// add a ticket to dataset
app.post('/home/admin/sell', server.sellTicket)
// UI way to remove an entrance
app.post('/home/admin/entered/undo', server.entranceUndo)
// Get who sold specified ticket
app.post('/home/admin/ticket/vendor', server.viewTicketVendor)

/* ===================*/

/* MANAGE PUBLIC RESOURCE */
// serve public files starting from root
app.use('/', express.static(path.join(__dirname, 'public'), { dotfiles: 'deny' }))
// let client to access to css and js library using specified path
app.use('/lib/pavilion', express.static(path.join(__dirname, 'node_modules/pavilion/dist'), { dotfiles: 'deny' }))

/* APPLICATION STARTUP */
app.listen(process.env.PORT || app.get('port'), () => {
  // debug info => initial message with listening port
  logger.info(`Website listening on port ${app.get('port')}`)
})
