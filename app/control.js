/*
 * Daniele Bissoli
 * FdP Tickets Manager - Server side logic
 * v0.0.2 - 2017-04-16
 */

'use strict'

const pino = require('pino')
const logger = pino({ level: process.env.LOG_LEVEL ?? 'info' })

const DB = require('./db.js')
const model = new DB(logger)

class Control {
  constructor() {}

  // login check
  isLoggedIn(request, response, next) {
    // if a user has already logged in,
    // then bring it directly to dashboard
    logger.debug({ session: request.session })
    if (request.session.cookie) {
      response.redirect('/home')
    } else {
      // otherwise go to login page
      return next()
    }
  }

  // check if user is currently logged in
  requireLogin(request, response, next) {
    // if it is not present a user, require to set one
    logger.debug({ session: request.session })
    if (!request.session.cookie) {
      response.redirect('/')
    } else {
      // otherwise go to the requested page
      return next()
    }
  }

  // check if current user has admin grant or not
  checkAdmin(request, response, next) {
    // here make a request to DB and get specific user
    logger.debug({ session: request.session })
    if (!request.session.user) {
      response.redirect('/')
    } else if (request.session.user === 'adm') {
      // otherwise go to the requested page
      return next()
    } else {
      response.redirect('/home')
    }
  }

  // login
  login(request, response) {
    // check if exist session data
    // to be used to initialize one
    if (request.session) {
      // shortcut for session information
      const sessionUser = request.session.user
      // check if username info has not already defined
      if (typeof (sessionUser) === 'undefined') {
        if (request.body) {
          // save username for later
          const usr = request.body.username

          // do the login
          model.login(usr, request.body.password)
            .then((result) => {
              // if login is  successful then set user session
              if (result === 0) {
                request.session.user = request.body.username
                logger.info({ user: request.session.user }, `user logged in`)
                // redirect user to main page
                response.redirect('/home')
              } else {
                // otherwise return an error to user
                // result = 1 => wrong pwd
                //        | 2 => undefined user
                // response.sendStatus(401);
                if (result === 1) {
                  logger.warn({ user: request.session.user }, `user wrong login`)
                } else if (result === 2) {
                  logger.warn({ user: request.session.user }, `user undefined`)
                }
                response.status(401)
                response.redirect('/')
              }
            })
            .catch((error) => {
              logger.error({ error }, 'failed to login')
              response.redirect('/')
            })
        } else {
          // notify user of bad input
          response.sendStatus(400)
        }
      }
      // keep track of who have done login for debug reasons
      logger.info({ sessionUser }, `user open app`)
    } else {
      response.redirect('/')
    }
  }

  // logout
  logout(request, response) {
    // save username for log purpose
    const { username } = request.session

    function callback(error) {
      // if it get some error track it,
      // otherwise send user to homepage
      if (error) {
        logger.warn({ user: username, error }, 'get an error at logging out user')
      } else {
        logger.info({ user: username }, 'user logged out')
      }
      response.redirect('/')
    }
    // do logout removing session
    request.session.destroy(callback)
  }

  // app endpoint
  startPage(request, response) {
    response.render('login', {})
  }

  // homepage
  home(request, response) {
    response.render('home', {})
  }

  // set a user entered
  entered(request, response) {
    const tmpNum = request.body.tck_num

    try {
      // check if is number and stop execution
      if (!tmpNum) { throw new Error('Ticket number not inserted!') }
      // check ticket number is really a number
      const tnum = parseInt(tmpNum)

      if (isNaN(tnum)) { throw new Error('Ticket inserted value is not valid!') }

      if (tnum < 1 || tnum > 1050) { throw new Error('Ticket number out of bound!') }

      model.setEntered1(tnum)
        .then((result) => {
          switch (result) {
          case 0:
            model.details(tnum)
              .then((attendee) => {
                logger.info(`Ticket ${attendee.ticket_num} entered at ${new Date()}`)
                response.render(
                  'checkTicket',
                  {
                    status: 0,
                    tnum: attendee.ticket_num,
                    ttype: attendee.ticket_type,
                    fname: attendee.first_name,
                    lname: attendee.last_name,
                  }
                )
              })
              .catch(logger.error(`Something wrong getting ${tnum} ticket details for enter update!`))
            break
          case 1:
            response.render('checkTicket', { status: result, tnum })
            break
          case 2:
            model.entered(tnum)
              .then(edate => {
                response.render('checkTicket', { status: result, tnum, date: new Date(edate) })
              })
              .catch((error) => { logger.error(error) })
            break
          default:
            response.render('checkTicket', { tnum: tmpNum })
          }
        })
        .catch(logger.warn('something wrong with entrance'))
    } catch (error) {
      logger.error({ error }, 'wrong or undefined ticket number')
      response.render('checkTicket', { tnum: (tmpNum === '' ? 'null' : tmpNum) })
    }
  }

  entered2(request, response) {
    response.location('/home')
    if (request.body.tnum) {
      try {
        const tnum = parseInt(request.body.tnum)

        if (isNaN(tnum)) { throw new Error('Ticket inserted value is not valid!') }

        if (tnum < 1 || tnum > 1050) { throw new Error('Ticket number out of bound!') }

        model.setEntered2(tnum)
          .then(result => {
            if (result) {
              response.render('home', { msg: 'Ticket committed successfully!' })
            } else {
              logger.error({ ticker: tnum }, `error committing ticket`)
              response.render('home', { msg: 'Error committing ticket!' })
            }
          })
          .catch(error => {
            logger.error({ error }, 'something wrong with entrance')
            response.render('home', { msg: 'Error committing ticket!' })
          })
      } catch (error) {
        logger.error({ error }, 'wrong or undefined ticket number')
        response.render('home', { msg: 'Error committing ticket!' })
      }
    }
  }

  getTickets(request, response) {
    Promise.all([model.list(), model.currentInside()])
      .then(result => {
        // get list of tickets
        response.render('tickets', { tickets: result[0], tentered: result[1] })
      })
      .catch((error) => { logger.error(error) })
  }

  ticketDetails(request, response) {
    // show specified ticket details
    const tmpNum = request.params.ticket_num
    try {
      const tnum = parseInt(tmpNum)

      if (isNaN(tnum)) { throw new Error('Ticket inserted value is not valid!') }

      model.details(tnum)
        .then((attendee) => {
          response.render('details', attendee)
        })
        .catch(logger.error(`Something went wrong getting ticket ${tnum} details!`))
    } catch (error) {
      logger.error({ error, ticket: tmpNum }, `wrong ticket number`)
      response.render('home', {})
    }
  }

  dashboard(request, response) {
    response.render('dashboard', {})
  }

  sellTicket(request, response) {
    if (request.body.tck_num) {
      try {
        const tnum = parseInt(request.body.tck_num)
        if (isNaN(tnum)) { throw new Error('Ticket inserted value is not valid!') }

        if (request.body.fname && request.body.lname) {
          model.sell(tnum, request.body.fname, request.body.lname)
            .then((status) => {
              // define which message display
              let info = ''
              if (status === 1) {
                info = 'Ticket Sold'
              } else if (status === 2) {
                info = 'Ticket not sold, since it has already sold!'
              } else {
                info = 'Update error!'
              }
              // notify the admin
              response.location('/home/admin/dashboard')
              response.render(
                'dashboard',
                {
                  msg: info,
                  status,
                }
              )
            })
            .catch(logger.error(`Something went wrong updating ticket ${tnum} details!`))
        } else {
          logger.warn('no ticket data provided!')
          response.location('/home/admin/dashboard')
          response.render('dashboard', { msg: 'No ticket data provided!' })
        }
      } catch (error) {
        logger.error({ error, ticket: request.body?.tck_num }, `wrong ticket number`)
        response.location('/home/admin/dashboard')
        response.render('checkTicket', { tnum: request.body?.tck_num })
      }
    } else {
      logger.warn('no ticket number provided for selling!')
      response.location('/home/admin/dashboard')
      response.render('dashboard', { msg: 'No ticket number provided for selling!' })
    }
  }

  entranceUndo(request, response) {
    if (request.body.tck_num) {
      try {
        const tnum = parseInt(request.body.tck_num)
        if (isNaN(tnum)) { throw new Error('Ticket inserted value is not valid!') }

        model.deleteEntrance(tnum)
          .then((status) => {
            response.location('/home/admin/dashboard')
            response.render(
              'dashboard',
              {
                msg: (status ? 'Entrance Updated' : 'Update error!'),
                status,
              }
            )
          })
          .catch(logger.error({ ticket: tnum }, `something went wrong updating entrance details`))
      } catch (error) {
        logger.error({ error, ticket: request.body?.tck_num }, `wrong ticket number`)
        response.location('/home/admin/dashboard')
        response.render('checkTicket', { tnum: request.body?.tck_num })
      }
    } else {
      logger.warn('No ticket number provided for selling!')
      response.location('/home/admin/dashboard')
      response.render('dashboard', { msg: 'No ticket number provided for selling!' })
    }
  }

  getTicketsInfo(request, response) {
    Promise.all([model.currentInside(), model.currentSold()])
      .then(result => {
        // get tickets info
        response.render('dashboard', { msg: `Entered: ${result[0]} - Sold ${result[1]}`, status: 1 })
      })
      .catch((error) => { logger.error(error) })
  }

  viewTicketVendor(request, response) {
    if (request.body.tck_num) {
      try {
        const tnum = parseInt(request.body.tck_num)
        if (isNaN(tnum)) { throw new Error('Ticket inserted value is not valid!') }

        model.checkVendor(tnum)
          .then((vendor) => {
            response.location('/home/admin/dashboard')
            response.render(
              'dashboard',
              {
                msg: (vendor ? vendor : 'No vendor!'),
                status: vendor,
              }
            )
          })
          .catch(logger.error(`Something went wrong updating entrance ${tnum} details!`))
      } catch (error) {
        logger.error({ error, ticket: request.body?.tck_num }, 'wrong ticket number')
        response.location('/home/admin/dashboard')
        response.render('checkTicket', { tnum: request.body?.tck_num })
      }
    } else {
      logger.warn('no ticket number provided for selling!')
      response.location('/home/admin/dashboard')
      response.render('dashboard', { msg: 'No ticket number provided for selling!' })
    }
  }
}

// export public server side functions
module.exports = Control
