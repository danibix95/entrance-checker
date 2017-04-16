/*
 * Daniele Bissoli
 * FdP Tickets Manager
 * v0.0.1 - 2017-04-16
 */

"use strict";

/* EXTERNAL LIBRARIES */
const express     = require("express");
const path        = require("path");
const bodyParser  = require("body-parser");
const session     = require("express-session");
const compression = require("compression");
const minify      = require("express-minify");

/* INTERNAL LIBRARIES */
const server = require("./server/server.js");
const logger = require("./server/logger.js");

/* APPLICATION INITALIZATION */
// specify process name, for eventually recognize it later
process.title = "fdp_tickets";
// process.env.NODE_ENV = "development";
// Initialize express application
const app = express();

// set some application information
app.set("title", "FdP Tickets Manager");
app.set("port", (process.env.PORT || 52017));
// tell to express to use pug engine
// template to render files
app.set('view engine', 'pug');

/* APPLICATION LOGIC */
// enable file compression
app.use(compression());
// enable minifying files
app.use(minify());
// enable application to parse encoded form
app.use(bodyParser.urlencoded({ extended: true }));

// enable session management
app.use(session({
  // string with hash session cookie
  secret: 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
  // don't save again a session if
  // it isn't modify from previous request
  resave: false,
  // don't save session for user
  // that haven't logged in
  saveUninitialized: false,
  // cookies options
  cookie: {
    sameSite : true,
    // need a certificate to only go over https
    secure: false,
    maxAge: 3600000,
  },
  // set session cookie name
  name: "nsession"
}));

/*===================*/
/* MANAGE API */
// these let user to skip login phase
// if it has already logged in
app.get("/", server.isLoggedIn);

// serve private default page
app.get("/home", server.requireLogin, server.home);

// set everything under /home/ path
// to be protected by unauthorized user
app.get("/home/*", server.requireLogin);

// list tickets status
app.get("/home/tickets", server.getTickets);
// get specific ticket status
app.get("/home/tickets/:num", server.checkTicket);

// set an user as entered to the event
app.post("/home/tickets/entered", server.entered);
// add a ticket to dataset
// app.post("/home/tickets/sell", server.checkAdmin, server.sellTicket);
// process login request checking user credentials
app.post("/login", server.login);
// process logout request
app.get("/logout", server.logout);
/*===================*/

/* MANAGE PUBLIC RESOURCE */
// serve public files starting from root
app.use("/", express.static(path.join(__dirname, "public"), {dotfiles: "deny"}));
// let client to access to css and js library using specified path
app.use('/lib/pavilion', express.static(path.join(__dirname, 'node_modules/pavilion/dist'), {dotfiles: "deny"}));

/* APPLICATION STARTUP */
app.listen(process.env.PORT || app.get("port"), function () {
  // debug info => initial message with listening port
  logger.info("Website listening on port " + app.get("port") + ".");
});