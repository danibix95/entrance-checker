/*
 * Daniele Bissoli
 * FdP Tickets Manager - Server side logic
 * v0.0.1 - 2017-04-16
 */

// internal libraries
const logger = require("./logger.js");
// library used to manage user data
const model = new (require("./db.js"))(); // new (..)() -> on the fly class construction

class Control {
  constructor() {}

  // login check
  isLoggedIn(request, response, next) {
    // if a user has already logged in,
    // then bring it directly to dashboard
    console.log(request.session.user);
    if (request.session.user) {
      response.redirect("/home");
    }
    // otherwise go to login page
    else {
      next();
    }
  }

  // check if user is currently logged in
  requireLogin(request, response, next) {
    // if it is not present a user, require to set one
    if (!request.session.user) {
      response.redirect('/');
    }
    // otherwise go to the requested page
    else {
      next();
    }
  }

  // check if current user has admin grant or not
  checkAdmin(request, response, next) {
    // here make a request to DB and get specific user
    if (!request.session.user) {
      response.redirect('/');
    }
    // otherwise go to the requested page
    else {
      if (request.session.user === "adm") {
        next();
      }
      else {
        response.redirect('/home');
      }
    }
  }

  //login
  login(request, response) {
    // check if exist session data
    // to be used to initialize one
    if (request.session) {
      // shortcut for session information
      var sessionUser = request.session.user;
      // check if username info has not already defined
      if (typeof (sessionUser) === "undefined") {
        if (request.body) {
          // save username for later
          var usr = request.body.username;

          // do the login
          model.login(usr, request.body.password)
              .then((result) => {
                // if login is  successful then set user session
                if (result === 0) {
                  request.session.user = usr;
                  logger.info(`User ${usr} logged in`)
                  // redirect user to main page
                  response.redirect("/home");
                }
                // otherwise return an error to user
                else {
                  // result = 1 => wrong pwd
                  //        | 2 => undefined user
                  // response.sendStatus(401);
                  if (result === 1) {
                    logger.warn(`User ${usr} wrong login`)
                  }
                  else if (result === 2 ) {
                    logger.warn(`User ${usr} undefined`)
                  }
                  response.status(401);
                  response.redirect("/");
                }
              })
              .catch((error) => {
                logger.log("Login error:\n", error);
                response.redirect("/");
              });
        }
        else {
          // notify user of bad input
          response.sendStatus(400);
        }
      }
      // keep track of who have done login for debug reasons
      logger.info(`User ${sessionUser} open app`);
    }
    else {
      response.redirect("/");
    }
  }

  // logout
  logout(request, response) {
    // save username for log purpose
    const username = request.session.username;

    function callback(error) {
      // if it get some error track it,
      // otherwise send user to homepage
      if (error) {
        logger.warn(`Get an error at ${new Date()} logging out user ${username}\n` + error);
      }
      else {
        logger.info(`User ${username} logged out at ${new Date()}`);
      }
      response.redirect('/');
    }
    // do logout removing session
    request.session.destroy(callback);
  }

  // app endpoint
  startPage(request, response) {
    response.render("login", {});
  }

  // homepage
  home(request, response) {
    response.render("home", {});
  }

  // set a user entered
  entered(request, response) {
    let tmpNum = request.body.tck_num;

    try {
      // check if is number and stop exectuion
      if (!tmpNum) throw new Error("Ticket number not inserted!");
      // check ticket number is really a number
      let tnum = parseInt(tmpNum);

      if (isNaN(tnum)) throw new Error("Ticket inserted value is not valid!");

      if (tnum < 1 || tnum > 1050) throw new Error("Ticket number out of bound!");

      model.setEntered1(tnum)
          .then((result) => {
            switch (result) {
              case 0:
                model.details(tnum)
                    .then((attendee) => {
                      logger.info(`Ticket ${attendee.ticket_num} entered at ${new Date()}`)
                      response.render(
                          "checkTicket",
                          {
                            status: 0,
                            tnum: attendee.ticket_num,
                            ttype: attendee.ticket_type,
                            fname: attendee.first_name,
                            lname: attendee.last_name
                          }
                      );
                    })
                    .catch(logger.error(`Something wrong getting ${tnum} ticket details for enter update!`));
                break;
              case 1:
                response.render("checkTicket", { status: result, tnum: tnum });
                break;
              case 2:
                model.entered(tnum)
                    .then(edate => {
                      response.render("checkTicket", { status: result, tnum: tnum, date: new Date(edate)});
                    })
                    .catch((error) => { logger.log("error", error); });
                break;
              default:
                response.render("checkTicket", { tnum : tmpNum});
            }
          })
          .catch(logger.warn("Something wrong with entrance!"));
    }
    catch (e) {
      logger.warn(`Wrong or undefined ticket number: ${e}`);
      response.render("checkTicket", { tnum : (tmpNum === "" ? "null" : tmpNum)})
    }
  }

  entered2(request, response) {
    response.location("/home");
    if (request.body.tnum) {
      try {
        let tnum = parseInt(request.body.tnum);

        if (isNaN(tnum)) throw new Error("Ticket inserted value is not valid!");

        if (tnum < 1 || tnum > 1050) throw new Error("Ticket number out of bound!");

        model.setEntered2(tnum)
            .then(result => {
              if (result) {
                response.render("home", { msg : "Ticket committed successfully!"})
              }
              else {
                logger.log("error", `Error committing ticket ${tnum}`);
                response.render("home", { msg : "Error committing ticket!"});
              }
            })
            .catch((error) => {
              logger.warn("Something wrong with entrance!")
              response.render("home", { msg : "Error committing ticket!"});
            });
      }
      catch (e) {
        logger.warn(`Wrong or undefined ticket number: ${e}`);
        response.render("home", { msg : "Error committing ticket!"});
      }
    }
  }

  getTickets(request, response) {
    Promise.all([model.list(), model.currentInside()])
        .then(result => {
          // get list of tickets
          response.render("tickets", { tickets: result[0], tentered: result[1] });
        })
        .catch((error) => { logger.log("error", error); });
  }

  ticketDetails(request, response) {
    // show specified ticket details
    let tmpNum = request.params.ticket_num;
    try {
      let tnum = parseInt(tmpNum);

      if (isNaN(tnum)) throw new Error("Ticket inserted value is not valid!");

      model.details(tnum)
          .then((attendee) => {
            response.render("details", attendee);
          })
          .catch(logger.error(`Something went wrong getting ticket ${tnum} details!`));
    }
    catch (e) {
      logger.warn(`Wrong ticket number (${tmpNum}): ${e}`);
      response.render("home", {})
    }
  }

  dashboard(request, response) {
    response.render("dashboard", {});
  }

  sellTicket(request, response) {
    if (request.body.tck_num) {
      try {
        let tnum = parseInt(request.body.tck_num);
        if (isNaN(tnum)) throw new Error("Ticket inserted value is not valid!");

        if (request.body.fname && request.body.lname) {
          model.sell(tnum, request.body.fname, request.body.lname)
              .then((status) => {
                response.location("/home/admin/dashboard");
                response.render(
                    "dashboard",
                    {
                      msg: (status ? "Ticket Sold" : "Update error!"),
                      status: status
                    }
                );
              })
              .catch(logger.error(`Something went wrong updating ticket ${tnum} details!`));
        }
        else {
          logger.warn("No ticket data provided!");
          response.location("/home/admin/dashboard");
          response.render("dashboard", { msg: "No ticket data provided!" });
        }
      }
      catch (e) {
        logger.warn(`Wrong ticket number (${tmpNum}): ${e}`);
        response.location("/home/admin/dashboard");
        response.render("checkTicket", { tnum : tmpNum});
      }
    }
    else {
      logger.warn("No ticket number provided for selling!");
      response.location("/home/admin/dashboard");
      response.render("dashboard", { msg: "No ticket number provided for selling!"});
    }
  }

  entranceUndo(request, response) {
    if (request.body.tck_num) {
      try {
        let tnum = parseInt(request.body.tck_num);
        if (isNaN(tnum)) throw new Error("Ticket inserted value is not valid!");

        model.deleteEntrance(tnum)
              .then((status) => {
                response.location("/home/admin/dashboard");
                response.render(
                    "dashboard",
                    {
                      msg: (status ? "Entrance Updated" : "Update error!"),
                      status: status
                    }
                );
              })
              .catch(logger.error(`Something went wrong updating entrance ${tnum} details!`));
      }
      catch (e) {
        logger.warn(`Wrong ticket number (${tmpNum}): ${e}`);
        response.location("/home/admin/dashboard");
        response.render("checkTicket", { tnum : tmpNum});
      }
    }
    else {
      logger.warn("No ticket number provided for selling!");
      response.location("/home/admin/dashboard");
      response.render("dashboard", { msg: "No ticket number provided for selling!"});
    }
  }

  viewTicketVendor(request, response) {
    if (request.body.tck_num) {
      try {
        let tnum = parseInt(request.body.tck_num);
        if (isNaN(tnum)) throw new Error("Ticket inserted value is not valid!");

        model.checkVendor(tnum)
              .then((vendor) => {
                response.location("/home/admin/dashboard");
                response.render(
                    "dashboard",
                    {
                      msg: (vendor ? vendor : "No vendor!"),
                      status: vendor
                    }
                );
              })
              .catch(logger.error(`Something went wrong updating entrance ${tnum} details!`));
      }
      catch (e) {
        logger.warn(`Wrong ticket number (${tmpNum}): ${e}`);
        response.location("/home/admin/dashboard");
        response.render("checkTicket", { tnum : tmpNum});
      }
    }
    else {
      logger.warn("No ticket number provided for selling!");
      response.location("/home/admin/dashboard");
      response.render("dashboard", { msg: "No ticket number provided for selling!"});
    }
  }
}

// export public server side functions
module.exports = Control;